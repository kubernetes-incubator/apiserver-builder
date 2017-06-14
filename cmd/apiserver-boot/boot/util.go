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
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"github.com/markbates/inflect"
)

var server string
var controllermanager string
var groupName string
var kindName string
var resourceName string
var versionName string
var copyright string
var domain string
var Repo string
var GoSrc string
var ignoreExists = false

// writeIfNotFound returns true if the file was created and false if it already exists
func writeIfNotFound(path, templateName, templateValue string, data interface{}) bool {
	// Make sure the directory exists
	exec.Command("mkdir", "-p", filepath.Dir(path)).CombinedOutput()

	// Don't create the doc.go if it exists
	if _, err := os.Stat(path); err == nil {
		return false
	} else if !os.IsNotExist(err) {
		log.Fatalf("Could not stat %s: %v", path, err)

	}
	create(path)

	t := template.Must(template.New(templateName).Funcs(
		template.FuncMap{
			"title": strings.Title,
			"lower": strings.ToLower,
			"plural": inflect.NewDefaultRuleset().Pluralize,
		},
	).Parse(templateValue))

	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		log.Fatalf("Failed to create %s: %v", path, err)
	}
	defer f.Close()

	err = t.Execute(f, data)
	if err != nil {
		log.Fatalf("Failed to create %s: %v", path, err)
	}

	return true
}

func getCopyright() string {
	if len(copyright) == 0 {
		// default to boilerplate.go.txt
		if _, err := os.Stat("boilerplate.go.txt"); err == nil {
			// Set this because it is passed to generators
			copyright = "boilerplate.go.txt"
			cr, err := ioutil.ReadFile(copyright)
			if err != nil {
				log.Fatalf("could not read copyright file %s", copyright)
			}
			return string(cr)
		}

		log.Fatalf("apiserver-boot create-resource requires the --copyright flag if boilerplate.go.txt does not exist")
	}

	if _, err := os.Stat(copyright); err != nil {
		if !os.IsNotExist(err) {
			log.Fatalf("Could not stat %s: %v", copyright, err)
		}
		return ""
	} else {
		cr, err := ioutil.ReadFile(copyright)
		if err != nil {
			log.Fatalf("could not read copyright file %s", copyright)
		}
		return string(cr)
	}
}

func create(path string) {
	f, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
}
