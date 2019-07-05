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

package generators

import (
	"io"
	"k8s.io/gengo/namer"
	"text/template"

	"k8s.io/gengo/generator"
)

type versionedGenerator struct {
	generator.DefaultGen
	apiversion *APIVersion
	apigroup   *APIGroup
}

var _ generator.Generator = &versionedGenerator{}

func CreateVersionedGenerator(apiversion *APIVersion, apigroup *APIGroup, filename string) generator.Generator {
	return &versionedGenerator{
		generator.DefaultGen{OptionalName: filename},
		apiversion,
		apigroup,
	}
}

func hasSubresources(version *APIVersion) bool {
	for _, v := range version.Resources {
		if len(v.Subresources) != 0 {
			return true
		}
	}
	return false
}

func (d *versionedGenerator) Imports(c *generator.Context) []string {
	imports := []string{
		"metav1 \"k8s.io/apimachinery/pkg/apis/meta/v1\"",
		"k8s.io/apimachinery/pkg/runtime",
		"github.com/kubernetes-incubator/apiserver-builder-alpha/pkg/builders",
		"k8s.io/apimachinery/pkg/runtime/schema",
		"k8s.io/apiserver/pkg/registry/generic",
		d.apigroup.Pkg.Path,
	}
	if hasSubresources(d.apiversion) {
		imports = append(imports, "k8s.io/apiserver/pkg/registry/rest")
	}

	return imports
}

func (d *versionedGenerator) Finalize(context *generator.Context, w io.Writer) error {
	temp := template.Must(template.New("versioned-template").Funcs(map[string]interface{}{
		"public": namer.IC,
	}).Parse(VersionedAPITemplate))
	return temp.Execute(w, d.apiversion)
}

var VersionedAPITemplate = `
func addKnownTypes(scheme *runtime.Scheme) error {
	// TODO this will get cleaned up with the scheme types are fixed
	scheme.AddKnownTypes(SchemeGroupVersion,
{{ range $api := .Resources -}}
		&{{ $api.Kind }}{},
		&{{ $api.Kind }}List{},
  {{ range $subresource := $api.Subresources -}}
		&{{ $subresource.Kind }}{},
  {{ end -}}
{{ end -}}
	)
	return nil
}

var (
	ApiVersion = builders.NewApiVersion("{{.Group}}.{{.Domain}}", "{{.Version}}").WithResources(
		{{ range $api := .Resources -}}
		{{$api.Group}}.{{$api.Group|public}}{{$api.Kind}}Storage,
		{{ if $api.REST }}{{ else -}}
		builders.NewApiResource( // Resource status endpoint
			{{ $api.Group }}.Internal{{ $api.Kind }}Status,
			func() runtime.Object { return &{{ $api.Kind }}{} },     // Register versioned resource
			func() runtime.Object { return &{{ $api.Kind }}List{} }, // Register versioned resource list
			&{{ $api.Group }}.{{ $api.StatusStrategy }}{builders.StatusStorageStrategySingleton},
		),{{ end -}}

		{{ range $subresource := $api.Subresources -}}
		builders.NewApiResourceWithStorage(
			{{ $api.Group }}.Internal{{ $subresource.Kind }}REST,
			func() runtime.Object { return &{{ $subresource.Request }}{} }, // Register versioned resource
			nil,
            {{ if $subresource.REST }}{{ $api.Group }}.New{{ $subresource.REST }}{{ else -}}
			func(generic.RESTOptionsGetter) rest.Storage { return &{{ $api.Group }}.{{ $subresource.Kind }}REST{ {{$api.Group}}.New{{$api.Kind}}Registry({{$api.Group}}.{{$api.Group|public}}{{$api.Kind}}Storage) } },
			{{ end -}}
		),
		{{ end -}}
		{{ end -}}
	)

	// Required by code generated by go2idl
	AddToScheme = (&runtime.SchemeBuilder{
		ApiVersion.SchemeBuilder.AddToScheme, 
		RegisterDefaults, 
		RegisterConversions,
		addKnownTypes,
		func(scheme *runtime.Scheme) error {
			metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
			return nil
		},
	}).AddToScheme

	SchemeBuilder = ApiVersion.SchemeBuilder
	localSchemeBuilder = SchemeBuilder
	SchemeGroupVersion = ApiVersion.GroupVersion
)

// Required by code generated by go2idl
// Kind takes an unqualified kind and returns a Group qualified GroupKind
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Required by code generated by go2idl
// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

{{ range $api := .Resources -}}
//
// {{.Kind}} Functions and Structs
//
// +k8s:deepcopy-gen=false
// Deprecated
type {{.Kind}}SchemeFns struct {
	builders.DefaultSchemeFns
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type {{$api.Kind}}List struct {
	metav1.TypeMeta ` + "`json:\",inline\"`" + `
	metav1.ListMeta ` + "`json:\"metadata,omitempty\"`" + `
	Items           []{{$api.Kind}} ` + "`json:\"items\"`" + `
}

{{ range $subresource := $api.Subresources -}}
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type {{$subresource.Request}}List struct {
	metav1.TypeMeta ` + "`json:\",inline\"`" + `
	metav1.ListMeta ` + "`json:\"metadata,omitempty\"`" + `
	Items           []{{$subresource.Request}} ` + "`json:\"items\"`" + `
}
{{ end }}{{ end -}}
`
