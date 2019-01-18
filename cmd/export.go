package cmd

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/snowdrop/query-api/pkg/helper/query"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"os"
)

type ExportOptions struct {
	args    []string
	genericclioptions.IOStreams
	component  *query.Component
	resources  *query.Resources
}

func NewExportOptions(streams genericclioptions.IOStreams) *ExportOptions {
	return &ExportOptions{
		IOStreams:   streams,
		component: query.NewComponent(),
		resources: query.NewResources(),
	}
}

type Params struct {
	output string
	component string
	resources string
	selector string
	ns string
}

func init() {
	var p = Params{}
	var exportExample = `
	# Collect the kubernetes resources of a component
	%[1]s export all
    `
	o := NewExportOptions(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	cmd := &cobra.Command{
		Use:          "export [flags]",
		Short:        "Collect kubernetes resources and export them",
		Example:      fmt.Sprintf(exportExample),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Validate(c, args); err != nil {
				return err
			}
			if err := o.Run(c, args, p); err != nil {
				return err
			}

			return nil
		},
	}
	cmd.PersistentFlags().StringVarP(&p.component, "component", "c","","Component to look for")
	cmd.PersistentFlags().StringVarP(&p.output, "output", "o","yaml","Output type : yaml, helm")
	rootCmd.AddCommand(cmd)
}

func (o *ExportOptions) Complete(cmd *cobra.Command, args []string) error {
	o.args = args
	return nil
}

func (o *ExportOptions) Validate(cmd *cobra.Command, args []string) error {
	return nil
}

func (o *ExportOptions) Run(cmd *cobra.Command, args []string, p Params) error {
	selector := "app=" + p.component
	infos, err := o.component.Query(selector)
	if err != nil {
		return err
	}

	if len(infos) > 0 {
		component := infos[0].Object.(metav1.Object)
		p.ns = component.GetNamespace()
		p.selector = "app=" + component.GetLabels()["app"]
		p.resources = "all,pvc"

		// Fetch the k8s resources which correspond to the component
		infos, err := o.resources.
			Query(p.selector, p.ns, p.resources)
		if err != nil {
			return err
		}

		// Generate the resources as Yaml list
		if p.output == "yaml" {
			o.resources.PrintYamlResult(infos)
		} else {
			// TODO - Generate Helm chart
		}

	} else {
		log.Errorf("No component found for %s",p.component)
	}

	return nil
}