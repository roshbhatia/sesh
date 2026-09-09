package cmd

import (
	"os"

	"github.com/roshbhatia/go-utils/ui"
	"github.com/roshbhatia/seshy/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show effective configuration",
	Long: `Show the effective configuration with the origin of each value.

Origin is one of env, file, or default, in that precedence. branchFormat is
additionally overridable per source repo through git config seshy.branchFormat,
which this global view does not read.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		settings, err := config.Effective()
		if err != nil {
			return err
		}
		rows := make([][]string, len(settings))
		for i, s := range settings {
			rows[i] = []string{s.Name, s.Value, ui.StdoutFaint(string(s.Origin))}
		}
		return table(os.Stdout, []string{"SETTING", "VALUE", "ORIGIN"}, rows)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
