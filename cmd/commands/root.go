package commands

import (
	"github.com/spf13/cobra"
	"os"
)

var configPath string

var rootCmd = &cobra.Command{
	Use:   "Web",
	Short: "Personal WebSite",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "./config.yml", "config path ./config.yml")
}
