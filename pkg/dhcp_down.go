package netlab

import (
	"fmt"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/sirupsen/logrus"
)

func DHCPDown(log *logrus.Entry, cfg *Config) error {
	runningDHCP := make(map[string]*process.Process)

	for vethName, veth := range cfg.Veths {
		names := GetVethNames(vethName)
		if veth.DHCP {
			runningDHCP[names[0]] = nil
		}

		for vid, vlan := range veth.VLANs {
			if vlan.DHCP {
				runningDHCP[fmt.Sprintf("%s.%d", names[0], vid)] = nil
			}
		}
	}

	processesList, err := process.Processes()
	if err != nil {
		log.WithError(err).Error("cannot list running processes")
	}

	ifaceList := make([]string, 0)

	for vethName, veth := range cfg.Veths {
		names := GetVethNames(vethName)
		if veth.DHCP {
			ifaceList = append(ifaceList, names[0])
		}

		for vid, vlan := range veth.VLANs {
			if vlan.DHCP {
				ifaceList = append(ifaceList, fmt.Sprintf("%s.%d", names[0], vid))
			}
		}
	}

	for _, p := range processesList {
		cmdLine, err := p.CmdlineSlice()
		if err != nil {
			log.WithError(err).Errorf("could not introspect process %d to extract its command line", p.Pid)
			return err
		}

		if cmdLine == nil && len(cmdLine) < 2 {
			continue
		}

		if cmdLine[0] == "dhclient" {
			iface := cmdLine[len(cmdLine)-1]
			if _, ok := runningDHCP[iface]; !ok {
				continue
			}
			log.Infof("found dhclient running on interface %s", iface)
			runningDHCP[iface] = p
		}
	}

	for _, iface := range ifaceList {
		if p, ok := runningDHCP[iface]; ok {
			if p == nil {
				continue
			}
			logrus.Infof("found dhclient running on %s (PID: %d), killing", iface, p.Pid)
			err = p.Kill()
			if err != nil {
				log.WithError(err).Errorf("could not kill dhclient/%d", p.Pid)
				return err
			}
		}
	}

	/*for vethName, veth := range cfg.Veths {
		names := GetVethNames(vethName)
		if p, ok := runningDHCP[names[0]]; ok {
			err = p.Kill()
			if err != nil {
				logrus.WithError(err).Error("failed to kill process")
			}
		}
	}*/

	return nil
}
