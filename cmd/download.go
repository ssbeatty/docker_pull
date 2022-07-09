package cmd

import (
	"docker_pull/dget"
	"github.com/spf13/cobra"
	"log"
	"sync"
)

var (
	proxy      string
	username   string
	password   string
	lsocks     bool
	lsocksPath string
)

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.PersistentFlags().StringVarP(&proxy, "proxy", "s", "", "First Select Proxy When Download Image")
	downloadCmd.PersistentFlags().StringVarP(&username, "username", "u", "", "User of Registry, Default From Docker Login & Only One Image Useful")
	downloadCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "Password of Registry, Default From Docker Login& Only One Image Useful")

	// proxy
	downloadCmd.PersistentFlags().BoolVarP(&lsocks, "lsocks", "", false, "Enable LightSockets")
	downloadCmd.PersistentFlags().StringVarP(&lsocksPath, "lsocks_path", "", "", "LightSockets Config Path, Default ~/.lightsocks.json")
}

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "download image",
	Args:  cobra.MinimumNArgs(1),
	Long:  `Download image to local.`,
	Run: func(cmd *cobra.Command, args []string) {
		startDownload(args)
	},
}

func startDownload(args []string) {
	// TODO other proxy client examples
	// https://github.com/gwuhaolin/lightsocks
	// ShadowSocksR or V2Ray
	client := dget.NewClient(&dget.Config{
		Proxy:   proxy,
		NeedBar: true,
		LightSock: dget.LightSock{
			Enable:     lsocks,
			ConfigPath: lsocksPath,
		},
	})

	wg := sync.WaitGroup{}

	for _, imgUri := range args {
		wg.Add(1)

		go func(imgUri string) {
			tag, err := client.ParseImageTag(imgUri)
			if err != nil {
				log.Printf("error when parse image uri: %s\n", imgUri)
				return
			}

			client.DownloadDockerImage(tag, username, password)

			defer wg.Done()
		}(imgUri)
	}
	wg.Wait()
}
