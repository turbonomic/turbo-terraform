provider "aws" {
  region = "us-east-1"
}

module "my_vpc" {
  source      = "../modules/vpc"
  vpc_cidr    = "192.168.0.0/16"
  tenancy     = "default"
  vpc_id      = "${module.my_vpc.vpc_id}"
  subnet_cidr = "192.168.1.0/24"
}

module "my_ec2" {
  source        = "../modules/ec2"
  ec2_count     = 3
  ami_id        = "ami-0915e09cc7ceee3ab"
  instance_type = "m5.large"
  subnet_id     = "${module.my_vpc.subnet_id}"
}

