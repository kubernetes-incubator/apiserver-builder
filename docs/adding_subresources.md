# Adding a subresource to a resource

## Create a resource

Create a resource under `pkg/apis/<group>/<version>/<resource>_types.go`

```go
// +resource:path=bars
// +subresource:request=Scale,path=scale,rest=ScaleBarREST
// +k8s:openapi-gen=true
type Bar struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BarSpec   `json:"spec,omitempty"`
	Status BarStatus `json:"status,omitempty"`
}

```

The following line tells the code generator to generate a subresource for this resource.

- under the path `bar/scale`
- with request Kind `Scale`
- implemented by the go type `ScaleBarREST`

Scale and ScaleBarREST live in the versioned package (same as the versioned resource definition)

```go
// +subresource:request=Scale,path=scale,rest=ScaleBarREST
```



## Create the subresource request

Define the request type in the same <kind>_types.go file

```go
// +genclient=true

// +subresource-request
type Scale struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Faculty int `json:"faculty,omitempty"`
}

```

Note the line:

```go
// +subresource-request
```

This tells the code generator that this is a subresource type and to
register it in the wiring.

## Create the REST implementation

Create the rest implementation in the *versioned* package.

Example:

```go
type ScaleBarREST struct {
	Registry BarRegistry
}

// Scale Subresource
var _ rest.CreaterUpdater = &ScaleBarREST{}
var _ rest.Patcher = &ScaleBarREST{}

func (r *ScaleBarREST) Create(ctx request.Context, obj runtime.Object) (runtime.Object, error) {
	scale := obj.(*Scale)
	b, err := r.Registry.GetBar(ctx, scale.Name, &metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
    // Do something with b...

    // Save the udpated b
	r.Registry.UpdateBar(ctx, b)
	return u, nil
}

// Get retrieves the object from the storage. It is required to support Patch.
func (r *ScaleBarREST) Get(ctx request.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	return nil, nil
}

// Update alters the status subset of an object.
func (r *ScaleBarREST) Update(ctx request.Context, name string, objInfo rest.UpdatedObjectInfo) (runtime.Object, bool, error) {
	return nil, false, nil
}

func (r *ScaleBarREST) New() runtime.Object {
	return &Scale{}
}

```


### Anatomy of a REST implementation

Define the struct type implementing the REST api.  The Registry
field is required, and provides a type safe library to read / write
instances of Bar from the storage.


```go
type ScaleBarREST struct {
	Registry BarRegistry
}
```


---

Enforce local compile time checks that the struct implements
the needed REST methods

```go
// Scale Subresource
var _ rest.CreaterUpdater = &ScaleBarREST{}
var _ rest.Patcher = &ScaleBarREST{}
```


---

Implement create and update methods using the Registry to update the parent resource.

```go
func (r *ScaleBarREST) Create(ctx request.Context, obj runtime.Object) (runtime.Object, error) {
    ...
}

// Update alters the status subset of an object.
func (r *ScaleBarREST) Update(ctx request.Context, name string, objInfo rest.UpdatedObjectInfo) (runtime.Object, bool, error) {
	...
}
```

---

Implement a read method using the Registry to read the parent resource.


```go
// Get retrieves the object from the storage. It is required to support Patch.
func (r *ScaleBarREST) Get(ctx request.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	...
}
```

---

Implement a method that creates new instance of the request.

```go
func (r *ScaleBarREST) New() runtime.Object {
	return &Scale{}
}
```


## Generate the code for your subresource

Run the code generation command to generate the wiring for your subresource.

`apiserver-boot generate`

## Invoke your subresource from a test

Use the RESTClient to call your subresource.  Client go is not generated
for subresources, so you will need to manually invoke the subresource.

```
client.RESTClient()
	err := restClient.Post().Namespace("default").
		Name("name").
		Resource("bars").
		SubResource("scale").
		Body(scale).Do().Error()
	...
```

