package query

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions/resource"
	"os"
)

type Component struct {}

func NewComponent() *Component {
	return &Component{}
}

func (c *Component) Config() *QueryOptions {
	q := NewQueryOptions()
	q.configFlags.ToRESTConfig()
	q.builder = resource.NewBuilder(q.configFlags)
	return q
}

func (c *Component) Query(selector string) ([]*resource.Info, error) {
	r := c.Config().builder.
		 Unstructured().
		 AllNamespaces(true).
		 ResourceTypeOrNameArgs(true, "component").
		 LabelSelector(selector).
		 Latest().
		 Flatten().
		 Do()

	infos, err := r.Infos()
	if err != nil {
		return nil, err
	}
	return infos, nil
}

func (c *Component) PrintYamlResult(infos []*resource.Info) {
	for _, info := range infos {
		//fmt.Printf("Type : %s, id: %s\n", info.Object.GetObjectKind().GroupVersionKind().Kind, info.Name)
		resource := info.Object
		if metadata, ok := resource.(metav1.Object); ok {
			//obj.SetCreationTimestamp(nt)
			metadata.SetGeneration(1)
			metadata.SetUID("")
			metadata.SetSelfLink("")
			metadata.SetCreationTimestamp(metav1.Time{})
			metadata.SetResourceVersion("")
			metadata.SetOwnerReferences(nil)
			metadata.SetNamespace("")
			metadata.SetAnnotations(map[string]string{})
		}
		// TODO : Is there a better way to access Spec, Status ...
		unstruct, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(resource)
		unstruct["status"] = nil

		// Convert Component to YAML output
		e := json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil)
		e.Encode(resource,os.Stdout)
	}
}

