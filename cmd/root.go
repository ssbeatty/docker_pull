package cmd

import (
	"github.com/spf13/cobra"
	"log"
)

var (
	proxy      string
	username   string
	password   string
	lsocks     bool
	lsocksPath string
	ssr        bool
	ssrPath    string
	ssrUrl     string
	useCache   bool
)

var rootCmd = &cobra.Command{
	Use:   "docker_pull",
	Short: "get a docker image",
	Long:  `get a docker image!`,
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(cleanCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
