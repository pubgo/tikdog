package tikdog_watcher

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pubgo/xerror"
)

func TestName(t *testing.T) {
	Start()

	d, _ := filepath.Abs(".")
	fmt.Println(d)

	go func() {
		for {
			info, _ := os.Lstat(d)
			fmt.Println(info.ModTime(), info.Sys())
			time.Sleep(time.Second)
		}
	}()

	xerror.Exit(Add(d, func(event interface{}) error {
		fmt.Println(event)
		return nil
	}))

	select {}
}
