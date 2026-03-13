package cmd

import (
	"fmt"

	"github.com/roshbhatia/sesh/internal/session"
	"github.com/spf13/cobra"
)

var pathCmd = &cobra.Command{
	Use:   "path <name>",
	Short: "Print session path",
	Long:  `Print the absolute path to a session.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		path, err := session.GetPath(name)
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
