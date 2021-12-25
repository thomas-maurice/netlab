package netlab

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

func Down(log *logrus.Entry, cfg *Config) error {
	err := cfg.Validate()
	if err != nil {
		return fmt.Errorf("could not create the lab: %w", err)
	}

	// delete bridges
	for bName := range cfg.Bridges {
		link, _ := netlink.LinkByName(bName)

		if link != nil {
			err := netlink.LinkDel(link)
			if err != nil {
				log.WithError(err).Errorf("could not remove interface %s", bName)
				return err
			}
		} else {
			logrus.Warningf("no such device %s, no action taken", bName)
		}
	}

	// delete veths

	for vethName := range cfg.Veths {
		names := GetVethNames(vethName)
		for _, name := range names {
			link, _ := netlink.LinkByName(names[0])

			if link != nil {
				err := netlink.LinkDel(link)
				if err != nil {
					log.WithError(err).Errorf("could not remove interface %s", name)
					return err
				}

			} else {
				logrus.Warningf("no such device %s, nothing to do", name)
			}
		}

	}
	return nil
}
