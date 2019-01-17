package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericclioptions/resource"
	"os"

)

var (
	infoExample = `
	# Collect the kubernetes resources of a component
	%[1]s export all
`
)

type ExportOptions struct {
	configFlags *genericclioptions.ConfigFlags
	builder *resource.Builder
	args    []string
	genericclioptions.IOStreams
}

func NewExportOptions(streams genericclioptions.IOStreams) *ExportOptions {
	return &ExportOptions{
		configFlags: genericclioptions.NewConfigFlags(),
		IOStreams:   streams,
	}
}

func NewCmdExport(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewExportOptions(streams)

	cmd := &cobra.Command{
		Use:          "export [flags]",
		Short:        "Collect kubernetes resources and export them",
		Example:      fmt.Sprintf(infoExample),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			if err := o.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}

func (o *ExportOptions) Complete(cmd *cobra.Command, args []string) error {
	o.args = args
	o.configFlags.ToRESTConfig()
	o.builder = resource.NewBuilder(o.configFlags)

	return nil
}

func (o *ExportOptions) Validate() error {
	return nil
}

func (o *ExportOptions) Run() error {
	r := o.builder.
		Unstructured().
		NamespaceParam("my-spring-boot").
		ResourceTypeOrNameArgs(true, o.args...).
		LabelSelector("app=my-spring-boot ").
		Latest().
		Flatten().
		Do()

	infos, err := r.Infos()
	if err != nil {
		panic(err)
	}

	o.PrintResult(infos)

	return nil
}

func (o *ExportOptions) PrintResult(infos []*resource.Info) {

	list := &metav1.List{
		TypeMeta: metav1.TypeMeta{
			Kind: "List",
			APIVersion: "v1",
		},
	}

	filters := []string{"Pod","ReplicationController","Component"}

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
			metadata.SetNamespace("")
			metadata.SetAnnotations(map[string]string{})
		}
		// TODO : Is there a better way to access Spec, Status ...
		unstruct, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(resource)
		unstruct["status"] = nil

		if kind == "Service" {
			spec := unstruct["spec"].(map[string]interface{})
			// Unset this field otherwise k8s will complaint
			spec["clusterIP"] = nil
		}

		if kind == "PersistentVolumeClaim" {
			spec := unstruct["spec"].(map[string]interface{})
			// Unset this field otherwise k8s will complaint
			spec["volumeName"] = nil
		}

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