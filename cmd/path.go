package cmd

import (
	"fmt"

	"github.com/roshbhatia/seshy/internal/session"
	"github.com/spf13/cobra"
)

var pathCmd = &cobra.Command{
	Use:               "path <name>",
	Short:             "Print session path",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeSessionNames,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := session.Resolve(args[0])
		if err != nil {
			return err
		}
		fmt.Println(path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pathCmd)
}
