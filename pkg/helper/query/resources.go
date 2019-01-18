package query

import (
	"bytes"
	"fmt"
	"github.com/snowdrop/query-api/pkg/helper"
	"github.com/snowdrop/query-api/pkg/helper/helm"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions/resource"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"os"
	"path/filepath"
	"strings"
)

var (
	filters = []string{"Pod", "ReplicationController", "Component"}
)

type Resources struct{}

func NewResources() *Resources {
	return &Resources{}
}

func (r *Resources) Config() *QueryOptions {
	q := NewQueryOptions()
	q.configFlags.ToRESTConfig()
	q.builder = resource.NewBuilder(q.configFlags)
	return q
}

func (r *Resources) Query(selector, ns, resources string) ([] *resource.Info, error) {
	resp := r.Config().builder.
		Unstructured().
		NamespaceParam(ns).
		ResourceTypeOrNameArgs(true, resources).
		LabelSelector(selector).
		Latest().
		Flatten().
		Do()

	infos, err := resp.Infos()
	if err != nil {
		return nil, err
	}

	return infos, nil
}

func (r *Resources) PrintYamlResult(infos []*resource.Info) {

	for _, info := range infos {
		//fmt.Printf("Type : %s, id: %s\n", info.Object.GetObjectKind().GroupVersionKind().Kind, info.Name)
		resource := info.Object
		kind := info.Object.GetObjectKind().GroupVersionKind().Kind
		if helper.In_Array(kind, filters) {
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
	e.Encode(list, os.Stdout)
}

func (r *Resources) GenerateHelmChart(component metav1.Object, infos []*resource.Info) error {
	name := component.GetName()
	version := "v1.0"

	fmt.Printf("Creating %s chart\n", name)
	chartname := filepath.Base(name)
	cfile := &chart.Metadata{
		Name:        chartname,
		Description: "Component Helm Chart " + chartname,
		Version:     version,
		AppVersion:  "1.0",
		ApiVersion:  helm.ApiVersionV1,
	}

	cpath := filepath.Dir(name)
	_, err := helm.Create(cfile, cpath)
	if err != nil {
		return err
	}

	for _, info := range infos {
		resource := info.Object
		kind := info.Object.GetObjectKind().GroupVersionKind().Kind
		if helper.In_Array(kind, filters) {
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

		var buf bytes.Buffer
		e := json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil)
		e.Encode(resource, &buf)

		tfile := &chart.Template{
			Name: "templates/" + strings.ToLower(kind) + ".yml",
			Data: buf.Bytes(),
		}

		_, err = helm.SaveTemplate(tfile, name)
		if err != nil {
			return err
		}
	}

	return nil
}
