package netlab

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

func Up(log *logrus.Entry, cfg *Config) error {
	err := cfg.Validate()
	if err != nil {
		return fmt.Errorf("could not create the lab: %w", err)
	}

	nlHandle, err := netlink.NewHandle(netlink.FAMILY_V4)
	if err != nil {
		logrus.WithError(err).Error("could not get a netlink handle")
		return err
	}

	defer nlHandle.Delete()

	// create bridges
	for bName := range cfg.Bridges {
		link, _ := netlink.LinkByName(bName)

		if link != nil {
			if link.Type() != "bridge" {
				log.Infof("link %s is not of type 'bridge', recreating", bName)
				err := netlink.LinkDel(link)
				if err != nil {
					log.WithError(err).Errorf("could not remove interface %s", bName)
					return err
				}
			}
		} else {
			logrus.Warningf("no such device %s, creating it", bName)
		}

		err := netlink.LinkAdd(&netlink.Bridge{LinkAttrs: netlink.LinkAttrs{Name: bName}})
		if err != nil && !os.IsExist(err) {
			log.WithError(err).Errorf("could not create bridge %s", bName)
			return err
		}

		link, _ = netlink.LinkByName(bName)
		if link == nil {
			logrus.WithError(err).Errorf("could not get a handle on %s", bName)
			return fmt.Errorf("could not get a handle on %s", bName)
		}
		if err := netlink.LinkSetUp(link); err != nil {
			logrus.WithError(err).Errorf("could bring bridge %s up", bName)
			return err
		}
	}

	// create veths

	for vethName, veth := range cfg.Veths {

		correctlySetup := true

		names := GetVethNames(vethName)
		for _, name := range names {
			link, _ := netlink.LinkByName(name)

			if link != nil {
				if link.Type() != "veth" {
					log.Infof("link %s is not of type 'veth' (%s), recreating", name, link.Type())
					err := netlink.LinkDel(link)
					if err != nil {
						log.WithError(err).Errorf("could not remove interface %s", name)
						return err
					}
					correctlySetup = false
				}
			} else {
				logrus.Warningf("no such device %s, needs creation", name)
				correctlySetup = false
			}
		}

		if !correctlySetup {
			if err := netlink.LinkAdd(&netlink.Veth{
				LinkAttrs: netlink.LinkAttrs{
					Name: names[0],
				},
				PeerName: names[1],
			}); err != nil {
				log.WithError(err).Errorf("could not create veth pair %s", vethName)
				return err
			}
		}

		// bringing them up
		for _, name := range names {
			link, _ := netlink.LinkByName(name)
			if link == nil {
				logrus.Errorf("veth interface does not exist: %s", name)
				return fmt.Errorf("no such interface %s", name)
			}

			if err := netlink.LinkSetUp(link); err != nil {
				logrus.WithError(err).Errorf("could bring veth %s up", name)
				return err
			}
		}

		// configure the slave end of the veth
		link, _ := netlink.LinkByName(names[1])
		if link == nil {
			logrus.Errorf("veth interface does not exist: %s", names[1])
			return fmt.Errorf("no such interface %s", names[1])
		}

		if veth.Master == "" {
			if link.Attrs().MasterIndex != 0 {
				log.Infof("removing master on %s", names[1])
				err = netlink.LinkSetNoMaster(link)
				if err != nil {
					logrus.WithError(err).Errorf("could not remove the master from %s", names[1])
					return fmt.Errorf("could not remove master from veth %s: %w", names[1], err)
				}
			}
		} else {
			masterBridge, _ := netlink.LinkByName(veth.Master)
			if masterBridge == nil {
				logrus.Errorf("veth interface does not exist: %s", veth.Master)
				return fmt.Errorf("no such interface %s", veth.Master)
			}

			err = netlink.LinkSetMaster(link, masterBridge)
			if err != nil {
				logrus.WithError(err).Errorf("could not set the master for %s", names[1])
				return fmt.Errorf("could not set the master bridge for veth %s: %w", names[1], err)
			}
		}

		vlanParent, _ := netlink.LinkByName(names[0])
		if vlanParent == nil {
			logrus.Errorf("veth interface does not exist: %s", names[0])
			return fmt.Errorf("no such interface %s", names[0])
		}

		// taking care of the VLANs
		for vid := range veth.VLANs {
			vlanName := fmt.Sprintf("%s.%d", names[0], vid)
			link, _ := netlink.LinkByName(vlanName)
			needsCreation := false
			if link != nil {
				if link.Type() != "vlan" {
					needsCreation = true
					log.Infof("link %s is not of type 'vlan' (%s), recreating", vlanName, link.Type())
					err := netlink.LinkDel(link)
					if err != nil {
						log.WithError(err).Errorf("could not remove interface %s", vlanName)
						return err
					}
				}
			} else {
				logrus.Warningf("no such device %s, needs creation", vlanName)
				needsCreation = true
			}

			if needsCreation {
				err = netlink.LinkAdd(&netlink.Vlan{
					VlanId:       vid,
					VlanProtocol: netlink.VLAN_PROTOCOL_8021Q,
					LinkAttrs: netlink.LinkAttrs{
						ParentIndex: vlanParent.Attrs().Index,
						Name:        vlanName,
					},
				})
				if err != nil {
					log.WithError(err).Errorf("could not create vlan interface %s", vlanName)
					return fmt.Errorf("could not create vlan interface %s, %w", vlanName, err)
				}
			}

			link, _ = netlink.LinkByName(vlanName)
			if link == nil {
				log.Errorf("veth interface does not exist: %s", vlanName)
				return fmt.Errorf("no such interface %s", vlanName)
			}

			if err := netlink.LinkSetUp(link); err != nil {
				log.WithError(err).Errorf("could bring veth %s up", vlanName)
				return err
			}
		}

		// purge potentially old vlans for the interface
		links, err := nlHandle.LinkList()
		if err != nil {
			log.WithError(err).Error("could not list interfaces")
			return err
		}

		for _, link := range links {
			if link.Attrs().ParentIndex == vlanParent.Attrs().Index && link.Type() == "vlan" {
				vlanIface := link.(*netlink.Vlan)
				if _, ok := veth.VLANs[vlanIface.VlanId]; !ok {
					log.Infof("vlan interface %s should not exist, deleting", link.Attrs().Name)
					err := netlink.LinkDel(link)
					if err != nil {
						log.WithError(err).Errorf("could not remove interface %s", link.Attrs().Name)
						return err
					}
				}
			}
		}

	}

	return nil
}
