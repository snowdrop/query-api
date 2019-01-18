package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	log "github.com/sirupsen/logrus"
	"github.com/pkg/errors"
	"os"
)

var rootCmd = &cobra.Command{Use: "odo"}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		checkError(err, "Root execution")
	}
}

// checkError prints the cause of the given error and exits the code with an
// exit code of 1.
// If the context is provided, then that is printed, if not, then the cause is
// detected using errors.Cause(err)
func checkError(err error, context string, a ...interface{}) {
	if err != nil {
		log.Debugf("Error:\n%v", err)
		if context == "" {
			fmt.Println(errors.Cause(err))
		} else {
			fmt.Printf(fmt.Sprintf("%s\n", context), a...)
		}
		os.Exit(1)
	}
}