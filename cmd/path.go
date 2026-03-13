package cmd

import (
	"fmt"

	"github.com/roshbhatia/sesh/internal/picker"
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

var selectCmd = &cobra.Command{
	Use:   "select",
	Short: "Interactively select session path",
	Long:  `Interactively select a session and print its path.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sessions, err := session.List()
		if err != nil {
			return fmt.Errorf("failed to list sessions: %w", err)
		}

		if len(sessions) == 0 {
			return fmt.Errorf("no sessions found")
		}

		names, descriptions := sessionPickerData(sessions)

		selected, err := picker.SelectOneWithDescription("Select session", names, descriptions)
		if err != nil {
			return err
		}

		path, err := session.GetPath(selected)
		if err != nil {
			return err
		}

		fmt.Println(path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pathCmd)
	rootCmd.AddCommand(selectCmd)
}
