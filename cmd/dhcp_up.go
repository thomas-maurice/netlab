package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	netlab "github.com/thomas-maurice/netlab/pkg"
)

var DHCPUpCmd = &cobra.Command{
	Use:   "dhcp-up",
	Short: "Launches DHCP clients",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := mustGetCfg(configFile)

		log := logrus.NewEntry(logrus.StandardLogger())

		err := netlab.DHCPUp(log, cfg)
		if err != nil {
			logrus.WithError(err).Fatal("could not run dhclients")
		}
	},
}

func InitDHCPUpCmd() {
}
