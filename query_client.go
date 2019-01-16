package main

import (
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericclioptions/resource"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
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
	for _, info := range infos {
		fmt.Printf("Type : %s, id: %s\n", info.Object.GetObjectKind().GroupVersionKind().Kind, info.Name)
	}
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
