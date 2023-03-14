package server

import (
	"github.com/spf13/cobra"
	"github.com/yin-zt/cmdb-notify/core/server"
)

// Cmd run http server
var Cmd = &cobra.Command{
	Use:   "server",
	Short: "Run cmdb notify server",
	Long:  `Run cmdb notify server`,
	Run: func(cmd *cobra.Command, args []string) {
		main()
	},
}

func main() {
	server.Start()
}
