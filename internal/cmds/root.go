package cmds

import (
	"github.com/pubgo/tikdog/internal/config"
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{}

func Run(cmd ...*cobra.Command) {
	rootCmd.AddCommand(cmd...)
	
	rootCmd.Use = config.Project
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error { return xerror.Wrap(cmd.Help()) }
	xerror.Exit(rootCmd.Execute())
}
