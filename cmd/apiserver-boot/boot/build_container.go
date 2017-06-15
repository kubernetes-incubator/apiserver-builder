/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package boot

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"io/ioutil"
)

var tag string

var createBuildContainerCmd = &cobra.Command{
	Use:   "build-container",
	Short: "Builds an container with the apiserver and controller-manager binaries",
	Long:  `Builds an container with the apiserver and controller-manager binaries`,
	Run:   RunBuildContainer,
}

func AddBuildContainer(cmd *cobra.Command) {
	cmd.AddCommand(createBuildContainerCmd)

	createBuildContainerCmd.Flags().StringVar(&tag, "tag", "", "use this tag for the image")
	createBuildContainerCmd.Flags().BoolVar(&generateForBuild, "generate", true, "if true, generate code before building")
}

func RunBuildContainer(cmd *cobra.Command, args []string) {
	if len(tag) == 0 {
		log.Fatalf("Must specify image tag using --tag when building containers")
	}

	dir, err := ioutil.TempDir(os.TempDir(), "apiserver-boot-build-container")
	if err != nil {
		log.Fatalf("failed to create temp directory %s %v", dir, err)
	}
	log.Printf("Will build docker image from directory %s", dir)

	log.Printf("Writing the Dockerfile.")

	path := filepath.Join(dir, "Dockerfile")
	writeIfNotFound(path, "dockerfile-template", dockerfileTemplate, dockerfileTemplateArguments{})

	log.Printf("Building binaries for linux amd64.")

	// Set the goos and goarch
	goos = "linux"
	goarch = "amd64"
	outputdir = dir
	RunBuild(cmd, args)

	log.Printf("Building the docker image.")

	c := exec.Command("docker", "build", "-t", tag, dir)
	fmt.Printf("%s\n", strings.Join(c.Args, " "))
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	err = c.Run()
	if err != nil {
		log.Fatal(err)
	}
}

type dockerfileTemplateArguments struct {
}

var dockerfileTemplate = `
FROM ubuntu:14.04

ADD apiserver .
ADD controller-manager .
`
