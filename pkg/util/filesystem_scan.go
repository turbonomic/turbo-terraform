package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func CreateTFStateToFilesMap(root, pattern string) (map[string]map[string]struct{}, error) {
	tfStateToFiles := make(map[string]map[string]struct{})
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched {
			dir := path[:strings.LastIndex(path, "/")+1]
			fileSet, exist := tfStateToFiles[path]
			if !exist {
				fileSet = make(map[string]struct{})
			}
			files, err := ioutil.ReadDir(dir)
			if err != nil {
				fmt.Println(err)
			}
			for _, f := range files {
				fileSet[dir+f.Name()] = struct{}{}
			}
			tfStateToFiles[path] = fileSet
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tfStateToFiles, nil
}
