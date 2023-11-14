package commands

import (
	"fmt"
	"github.com/mar-coding/personalWebsiteBackend/configs"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run Personal WebSite",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := configs.NewConfig(configPath)
		if err != nil {
			return err
		}

		fmt.Println(cfg.ExtraData.Email)

		return nil
	},
}

func permissionOptions(methodFullName string) ([]int32, bool, bool, bool, error) {
	return []int32{}, false, false, false, nil
}
