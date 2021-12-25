package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	netlab "github.com/thomas-maurice/netlab/pkg"
)

var DHCPDownCmd = &cobra.Command{
	Use:   "dhcp-down",
	Short: "Stops DHCP clients",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := mustGetCfg(configFile)

		log := logrus.NewEntry(logrus.StandardLogger())

		err := netlab.DHCPDown(log, cfg)
		if err != nil {
			logrus.WithError(err).Fatal("could not stop dhclients")
		}
	},
}

func InitDHCPDownCmd() {
}
