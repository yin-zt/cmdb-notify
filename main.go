package main

import (
	"github.com/spf13/cobra"
	"github.com/yin-zt/cmdb-notify/cmd/server"
)

var (
	VERSION    string
	BUILD_TIME string
	GO_VERSION string
)

func main() {
	root := cobra.Command{
		Use:   "cmdb-notify",
		Short: "A tool to analy cmdb notify data",
	}
	root.AddCommand(
		server.Cmd,
	)
	root.Execute()
}
