package main

import (
	"flag"
	"fmt"
	"github.com/enlinxu/turbo-terraform/pkg/discovery"
	"github.com/enlinxu/turbo-terraform/pkg/registration"
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-go-sdk/pkg/probe"
	"github.com/turbonomic/turbo-go-sdk/pkg/service"
	"os"
	"os/signal"
	"syscall"
)

type disconnectFromTurboFunc func()

var (
	targetConf string
	opsMgrConf string
	tfPath     string
)

func getFlags() {
	flag.StringVar(&opsMgrConf, "turboConf", "./conf/turbo.json", "configuration file of OpsMgr")
	flag.StringVar(&targetConf, "targetConf", "./conf/target.json", "configuration file of target")
	flag.StringVar(&tfPath, "tfpath", "", "Terraform Assets Location")
	flag.Set("alsologtostderr", "true")
	flag.Parse()
}

func buildProbe(targetConf string, stop chan struct{}) (*probe.ProbeBuilder, error) {
	config, err := discovery.NewTargetConf(targetConf)
	if err != nil {
		return nil, fmt.Errorf("failed to load json conf:%v", err)
	}

	regClient := &registration.TFRegistrationClient{}
	discoveryClient := discovery.NewDiscoveryClient(config, &tfPath)

	builder := probe.NewProbeBuilder(config.TargetType, config.ProbeCategory).
		RegisteredBy(regClient).
		WithActionPolicies(regClient).
		WithEntityMetadata(regClient).
		DiscoversTarget(config.Identifier, discoveryClient)

	return builder, nil
}

func createTapService() (*service.TAPService, error) {
	turboConfig, err := service.ParseTurboCommunicationConfig(opsMgrConf)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpsMgrConfig: %v", err)
	}

	stop := make(chan struct{})
	probeBuilder, err := buildProbe(targetConf, stop)
	if err != nil {
		return nil, fmt.Errorf("failed to create probe: %v", err)
	}

	tapService, err := service.NewTAPServiceBuilder().
		WithTurboCommunicator(turboConfig).
		WithTurboProbe(probeBuilder).
		Create()

	if err != nil {
		return nil, fmt.Errorf("error when creating TapService: %v", err.Error())
	}

	return tapService, nil
}

func main() {
	getFlags()
	glog.Info("Starting Turbo Terraform...")
	glog.Infof("GIT_COMMIT: %s", os.Getenv("GIT_COMMIT"))
	if tfPath == "" {
		glog.Errorf("Terraform file path has to be provided with -tfpath=<file_path>")
		os.Exit(1)
	}
	tap, err := createTapService()
	if err != nil {
		glog.Errorf("failed to create tapServier: %v", err)
	}
	handleExit(func() { tap.DisconnectFromTurbo() })

	tap.ConnectToTurbo()

}

// handleExit disconnects the tap service from Turbo service when Terraform probe is terminated
func handleExit(disconnectFunc disconnectFromTurboFunc) {
	glog.V(4).Infof("*** Handling Turbo Terraform Termination ***")
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGHUP)

	go func() {
		select {
		case sig := <-sigChan:
			glog.V(2).Infof("Signal %s received. Disconnecting Turbo Terraform from Turbo server...\n", sig)
			disconnectFunc()
		}
	}()
}
