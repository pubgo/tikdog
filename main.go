package main

import (
	"github.com/pubgo/tikdog/internal/cmds"
	"github.com/pubgo/tikdog/tikdog_sync"
)

func main() {
	cmds.Run(
		tikdog_sync.GetCmd(),
	)
}
