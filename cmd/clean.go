package cmd

import (
	"docker_pull/dget"
	"github.com/spf13/cobra"
	"log"
)

var (
	cleanAll bool
)

func init() {
	cleanCmd.PersistentFlags().BoolVarP(&cleanAll, "all", "a", false, "Clean All Cache, Example config.")
}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "clean cache",
	Args:  cobra.MinimumNArgs(0),
	Long:  `clean cache`,
	Run: func(cmd *cobra.Command, args []string) {
		startCleanCmd(args)
	},
}

func startCleanCmd(args []string) {
	err := dget.CleanCache(cleanAll)
	if err != nil {
		log.Printf("error when clean cache file\n")
		return
	}
}
