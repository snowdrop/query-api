package cmd

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/snowdrop/query-api/pkg/helper/query"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"os"
)

type SearchOptions struct {
	args []string
	genericclioptions.IOStreams
	component *query.Component
	resources *query.Resources
}

func NewSearchOptions(streams genericclioptions.IOStreams) *SearchOptions {
	return &SearchOptions{
		IOStreams: streams,
		component: query.NewComponent(),
		resources: query.NewResources(),
	}
}

type SearchParams struct {
	resources string
	selector  string
	ns        string
}

func init() {
	var p = SearchParams{}
	var searchExample = `
	# Search about kubernetes resources using a field selector
	%[1]s search all
    `
	o := NewSearchOptions(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	cmd := &cobra.Command{
		Use:          "search [flags]",
		Short:        "Search about kubernetes resources using a field selector",
		Example:      fmt.Sprintf(searchExample),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Validate(c, args); err != nil {
				return err
			}
			if err := o.Search(c, args, p); err != nil {
				return err
			}

			return nil
		},
	}
	cmd.PersistentFlags().StringVarP(&p.selector, "FieldSelector", "f", "", "FieldSelector to look for")
	cmd.PersistentFlags().StringVarP(&p.resources, "Kubernetes resources", "r", "", "Kubernetes resources to query")
	cmd.PersistentFlags().StringVarP(&p.ns, "Namespace", "n", "", "Nameapsce to query")
	rootCmd.AddCommand(cmd)
}

func (o *SearchOptions) Complete(cmd *cobra.Command, args []string) error {
	o.args = args
	return nil
}

func (o *SearchOptions) Validate(cmd *cobra.Command, args []string) error {
	return nil
}

func (o *SearchOptions) Search(cmd *cobra.Command, args []string, p SearchParams) error {

	// Fetch the k8s resources which correspond to the component
	infos, err := o.resources.
		QueryByField(p.selector, p.ns, p.resources)
	if err != nil {
		return err
	}

	// Print the result
	log.Info("Result", infos)

	return nil
}
