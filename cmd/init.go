package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize sesh",
	Long:  `Initialize sesh (placeholder for future setup).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("sesh initialized")
		return nil
	},
	Hidden: true,
}

func init() {
	rootCmd.AddCommand(initCmd)
}
