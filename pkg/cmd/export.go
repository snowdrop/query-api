package cmd

import (
	"fmt"
	"github.com/snowdrop/query-api/pkg/helper/query"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"strings"
)

var (
	exportExample = `
	# Collect the kubernetes resources of a component
	%[1]s export all
`
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

func NewCmdExport(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewExportOptions(streams)

	cmd := &cobra.Command{
		Use:          "export [flags]",
		Short:        "Collect kubernetes resources and export them",
		Example:      fmt.Sprintf(exportExample),
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
	return nil
}

func (o *ExportOptions) Validate() error {
	return nil
}

func (o *ExportOptions) Run() error {
	resources := strings.Join(o.args[:],",")
	selector := "app=my-spring-boot"
	ns := "my-spring-boot"

	o.component.
		Query(selector)

	o.resources.
		Query(selector, ns, resources)



	return nil
}