package cmds

import (
	"github.com/pubgo/tikdog/internal/config"
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"
)

var rootCmd = &cobra.Command{}

func Run() {
	rootCmd.Use = config.Project
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		return xerror.Wrap(cmd.Help())
	}
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		xerror.Panic(viper.BindPFlags(cmd.Flags()))

		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		return nil
	}
}
