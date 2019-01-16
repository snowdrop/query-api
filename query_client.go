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

	objs := make([]runtime.Object, len(infos))
	for ix := range infos {
		table, err := DecodeIntoTable(infos[ix].Object)
		if err == nil {
			infos[ix].Object = table
		} else {
			// if we are unable to decode server response into a v1beta1.Table,
			// fallback to client-side printing with whatever info the server returned.
			fmt.Printf("Unable to decode server response into a Table. Falling back to hardcoded types: %v", err)
		}
		objs[ix] = infos[ix].Object
	}

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

func DecodeIntoTable(obj runtime.Object) (runtime.Object, error) {
	if obj.GetObjectKind().GroupVersionKind().Kind != "Table" {
		return nil, fmt.Errorf("attempt to decode non-Table object into a v1beta1.Table")
	}

	unstr, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("attempt to decode non-Unstructured object")
	}
	table := &metav1beta1.Table{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstr.Object, table); err != nil {
		return nil, err
	}

	for i := range table.Rows {
		row := &table.Rows[i]
		if row.Object.Raw == nil || row.Object.Object != nil {
			continue
		}
		converted, err := runtime.Decode(unstructured.UnstructuredJSONScheme, row.Object.Raw)
		if err != nil {
			return nil, err
		}
		row.Object.Object = converted
	}

	return table, nil
}
