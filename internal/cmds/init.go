package cmds

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/pubgo/tikdog/internal/config"
	"github.com/pubgo/tikdog/tikdog_util"
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use: "init",
		Run: func(cmd *cobra.Command, args []string) {
			home := filepath.Join(xerror.PanicStr(homedir.Dir()), "."+config.Project, "config")
			xerror.Panic(os.MkdirAll(home, 0755))

			fmt.Println("config home:", home)

			cfgPath := filepath.Join(home, "config.yaml")
			if !tikdog_util.IsNotExist(cfgPath) {
				return
			}

			xerror.Panic(ioutil.WriteFile(cfgPath, []byte(config.GetDefault()), 0600))
		},
	})
}
