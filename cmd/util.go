package cmd

import (
	"github.com/sirupsen/logrus"
	netlab "github.com/thomas-maurice/netlab/pkg"
)

func mustGetCfg(fileName string) *netlab.Config {
	cfg, err := netlab.LoadConfigFromFile(fileName)
	if err != nil {
		logrus.WithError(err).Fatalf("could not load configuration %s", fileName)
	}

	return cfg
}
