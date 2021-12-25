package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	netlab "github.com/thomas-maurice/netlab/pkg"
)

var DownCmd = &cobra.Command{
	Use:   "down",
	Short: "Creates and updates the infra as trequired",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := mustGetCfg(configFile)

		log := logrus.NewEntry(logrus.StandardLogger())

		err := netlab.Down(log, cfg)
		if err != nil {
			logrus.WithError(err).Fatal("could not delete infra")
		}
	},
}

func InitDownCmd() {
}
