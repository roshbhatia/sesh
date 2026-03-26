package cmd

import (
	"fmt"

	"github.com/roshbhatia/seshy/internal/session"
	"github.com/roshbhatia/seshy/internal/ui"
	"github.com/spf13/cobra"
)

var pathCmd = &cobra.Command{
	Use:   "path <name>",
	Short: "Print session path",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if !session.Exists(name) {
			return fmt.Errorf("session %s not found", ui.AccentBold.Render(name))
		}
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
