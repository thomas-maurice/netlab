package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	netlab "github.com/thomas-maurice/netlab/pkg"
)

var UpCmd = &cobra.Command{
	Use:   "up",
	Short: "Creates and updates the infra as trequired",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := mustGetCfg(configFile)

		log := logrus.NewEntry(logrus.StandardLogger())

		err := netlab.Up(log, cfg)
		if err != nil {
			logrus.WithError(err).Fatal("could not create infra")
		}
	},
}

func InitUpCmd() {
}
