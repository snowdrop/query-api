package query

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericclioptions/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	list = &metav1.List{
		TypeMeta: metav1.TypeMeta{
			Kind: "List",
			APIVersion: "v1",
		},
	}
)

type QueryOptions struct {
	builder *resource.Builder
	configFlags *genericclioptions.ConfigFlags
}

func NewQueryOptions() *QueryOptions {
	return &QueryOptions{
		configFlags: genericclioptions.NewConfigFlags(),
	}
}
