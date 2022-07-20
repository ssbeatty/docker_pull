package cmd

import (
	"docker_pull/dget"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

func init() {
	pullCmd.PersistentFlags().StringVarP(&proxy, "proxy", "s", "", "First Select Proxy When Download Image")
	pullCmd.PersistentFlags().StringVarP(&username, "username", "u", "", "User of Registry, Default From Docker Login & Only One Image Useful")
	pullCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "Password of Registry, Default From Docker Login& Only One Image Useful")
	pullCmd.PersistentFlags().BoolVarP(&useCache, "cache", "", false, "Use Cache")

	// lightsocks
	pullCmd.PersistentFlags().BoolVarP(&lsocks, "lsocks", "", false, "Enable LightSockets")
	pullCmd.PersistentFlags().StringVarP(&lsocksPath, "lsocks_path", "", "", "LightSockets Config Path, Default ~/.lightsocks.json")
	// ssr
	pullCmd.PersistentFlags().BoolVarP(&ssr, "ssr", "", false, "Enable SSR")
	pullCmd.PersistentFlags().StringVarP(&ssrPath, "ssr_path", "", "", "SSR Config Path, Default ~/.shadowsocks.json")
	pullCmd.PersistentFlags().StringVarP(&ssrUrl, "ssr_url", "", "", "SSR Base64 URL, Start With ssr:// , Will Write To ~/.shadowsocks.json")
}

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "pull image",
	Args:  cobra.MinimumNArgs(1),
	Long:  `Download image to local and load use Docker`,
	Run: func(cmd *cobra.Command, args []string) {
		startPullCmd(args)
	},
}

func dockerLoad(path string) error {
	_, err := exec.LookPath("docker")
	if err != nil {
		return err
	}

	cmd := exec.Command("docker", "load", "-i", path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func startPullCmd(args []string) {
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
			dockerTar := strings.ReplaceAll(tag.Repo, "/", "_") + "_" + tag.Img + "_" + tag.Tag + ".tar"

			err = dockerLoad(dockerTar)
			if err != nil {
				log.Printf("error when docker load %s, err: %v\n", dockerTar, err)
				return
			}
			_ = os.Remove(dockerTar)

		}(imgUri)
	}
	wg.Wait()
}
