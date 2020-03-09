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

package main

import (
	"os"
	"runtime"
	"flag"

	"k8s.io/gengo/args"
	"github.com/spf13/pflag"
	"k8s.io/klog"
	"sigs.k8s.io/apiserver-builder-alpha/cmd/apiregister-gen/generators"
)

func main() {

	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	arguments := args.Default()

	// Override defaults.
	arguments.OutputFileBaseName = "zz_generated.api.register"

	// Custom args.
	fs := &flag.FlagSet{}
	klog.InitFlags(fs)
	pfs := &pflag.FlagSet{}
	pfs.AddGoFlagSet(fs)

	customArgs := &generators.CustomArgs{}
	arguments.CustomArgs = customArgs
	arguments.AddFlags(pfs)

	g := generators.Gen{}
	if err := g.Execute(arguments); err != nil {
		klog.Fatalf("Error: %v", err)
	}
	klog.V(2).Info("Completed successfully.")
}
