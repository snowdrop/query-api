package main

import (
	"github.com/snowdrop/query-api/pkg/cmd"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"os"
)

func main() {
	var rootCmd = &cobra.Command{Use: "odo"}
	rootCmd.AddCommand(cmd.NewCmdExport(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}))
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}