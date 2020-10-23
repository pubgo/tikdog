package cmds

import (
	"github.com/pubgo/tikdog/internal/config"
	"github.com/pubgo/tikdog/internal/tikdog"
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{}

func Run() {
	rootCmd.Use = config.Project
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		return xerror.Wrap(tikdog.New().Run())
	}
	xerror.Exit(rootCmd.Execute())
}
