package main

import (
	// "fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericclioptions/resource"
	"os"
)

var (
	scheme = runtime.NewScheme()
)

func main() {
	o := NewParams(genericclioptions.ConfigFlags{})
	o.configFlags.ToRESTConfig()
	o.builder = resource.NewBuilder(o.configFlags)

	r := o.builder.
		Unstructured().
		NamespaceParam("my-spring-boot").
		ResourceTypeOrNameArgs(true, "all").
		LabelSelector("app=my-spring-boot ").
		Latest().
		Flatten().
		Do()

	infos, err := r.Infos()
	if err != nil {
		panic(err)
	}

	PrintResult(infos)
}


func PrintResult(infos []*resource.Info) {

	list := &metav1.List{
		TypeMeta: metav1.TypeMeta{
			Kind: "List",
			APIVersion: "v1",
		},
	}

	filters := []string{"Pod","ReplicationController"}

	for _, info := range infos {
		//fmt.Printf("Type : %s, id: %s\n", info.Object.GetObjectKind().GroupVersionKind().Kind, info.Name)
		resource := info.Object
		kind := info.Object.GetObjectKind().GroupVersionKind().Kind
		if in_array(kind, filters) {
			continue
		}
		if metadata, ok := resource.(metav1.Object); ok {
			//obj.SetCreationTimestamp(nt)
			metadata.SetGeneration(1)
			metadata.SetUID("")
			metadata.SetSelfLink("")
			metadata.SetCreationTimestamp(metav1.Time{})
			metadata.SetResourceVersion("")
			metadata.SetOwnerReferences(nil)
		}
		unstruct, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(resource)
		unstruct["status"] = nil
		// e := json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil)
		// e.Encode(resource,os.Stdout)
		list.Items = append(list.Items, runtime.RawExtension{Object: resource.DeepCopyObject()})
	}
	// Convert List of objects to YAML list
	e := json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil)
	e.Encode(list,os.Stdout)
}

func in_array(val string, resources []string) bool {
	for _, r := range resources {
		if r == val {
			return true
		}
	}
	return false
}

type Params struct {
	configFlags *genericclioptions.ConfigFlags
	builder     *resource.Builder
}

func NewParams(flags genericclioptions.ConfigFlags) *Params {
	return &Params{
		configFlags: genericclioptions.NewConfigFlags(),
	}
}
