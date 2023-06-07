package cmd

import (
	"fmt"
	"github.com/lijiang2014/cwl.go/frontend/server"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run cwl.go as a server",
	Long:  `cwl.go can run as a grpc server. Use any grpc client to connect and control it.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := server.StartServe(4321, "", "")
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
