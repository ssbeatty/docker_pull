package cmd

import (
	"docker_pull/dget"
	"github.com/spf13/cobra"
	"log"
	"sync"
)

func init() {
	downloadCmd.PersistentFlags().StringVarP(&proxy, "proxy", "s", "", "First Select Proxy When Download Image")
	downloadCmd.PersistentFlags().StringVarP(&username, "username", "u", "", "User of Registry, Default From Docker Login & Only One Image Useful")
	downloadCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "Password of Registry, Default From Docker Login& Only One Image Useful")
	downloadCmd.PersistentFlags().BoolVarP(&useCache, "cache", "", false, "Use Cache")

	// lightsocks
	downloadCmd.PersistentFlags().BoolVarP(&lsocks, "lsocks", "", false, "Enable LightSockets")
	downloadCmd.PersistentFlags().StringVarP(&lsocksPath, "lsocks_path", "", "", "LightSockets Config Path, Default ~/.lightsocks.json")
	// ssr
	downloadCmd.PersistentFlags().BoolVarP(&ssr, "ssr", "", false, "Enable SSR")
	downloadCmd.PersistentFlags().StringVarP(&ssrPath, "ssr_path", "", "", "SSR Config Path, Default ~/.shadowsocks.json")
	downloadCmd.PersistentFlags().StringVarP(&ssrUrl, "ssr_url", "", "", "SSR Base64 URL, Start With ssr:// , Will Write To ~/.shadowsocks.json")
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
	client := dget.NewClient(&dget.Config{
		Proxy:    proxy,
		NeedBar:  true,
		UseCache: useCache,
		LightSock: dget.LightSock{
			Enable:     lsocks,
			ConfigPath: lsocksPath,
		},
		SSR: dget.SSR{
			Enable:     ssr,
			ConfigPath: ssrPath,
			Url:        ssrUrl,
		},
	})

	wg := sync.WaitGroup{}

	for _, imgUri := range args {
		wg.Add(1)

		go func(imgUri string) {
			defer wg.Done()

			tag, err := client.ParseImageTag(imgUri)
			if err != nil {
				log.Printf("error when parse image uri: %s\n", imgUri)
				return
			}

			client.DownloadDockerImage(tag, username, password)

		}(imgUri)
	}
	wg.Wait()
}
