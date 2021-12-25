package netlab

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

func DHCPUp(log *logrus.Entry, cfg *Config) error {
	nlHandle, err := netlink.NewHandle(netlink.FAMILY_V4)
	if err != nil {
		return err
	}
	defer nlHandle.Delete()

	neededDHCP := make(map[string]*process.Process)

	for vethName, veth := range cfg.Veths {
		names := GetVethNames(vethName)
		if veth.DHCP {
			neededDHCP[names[0]] = nil
		}

		for vid, vlan := range veth.VLANs {
			if vlan.DHCP {
				neededDHCP[fmt.Sprintf("%s.%d", names[0], vid)] = nil
			}
		}
	}

	processesList, err := process.Processes()
	if err != nil {
		log.WithError(err).Error("cannot list running processes")
	}

	for _, p := range processesList {
		cmdLine, err := p.CmdlineSlice()
		if err != nil {
			log.WithError(err).Error("could not introspect process %d to extract its command line", p.Pid)
			return err
		}

		if cmdLine == nil && len(cmdLine) < 2 {
			continue
		}

		if cmdLine[0] == "dhclient" {
			iface := cmdLine[len(cmdLine)-1]
			if _, ok := neededDHCP[iface]; !ok {
				continue
			}
			log.Infof("found dhclient running on interface %s", iface)
			neededDHCP[iface] = p
		}
	}

	for ifaceName, runningProcess := range neededDHCP {
		logrus.Infof("processing DHCP for %s", ifaceName)
		if runningProcess != nil {
			logrus.Infof("dhclient is already running on %s (PID: %d)", ifaceName, runningProcess.Pid)
			continue
		}

		cmd := exec.Command(
			"dhclient",
			"-4",
			"-v",
			ifaceName,
		)

		stderr, _ := cmd.StderrPipe()
		stdout, _ := cmd.StdoutPipe()
		err = cmd.Start()
		if err != nil {
			return err
		}

		log.Infof("started dhclient for %s on pid %d", ifaceName, cmd.Process.Pid)

		stderrScanner := bufio.NewScanner(stderr)
		stderrScanner.Split(bufio.ScanLines)
		stdoutScanner := bufio.NewScanner(stdout)
		stdoutScanner.Split(bufio.ScanLines)

		go func() {
			for stderrScanner.Scan() {
				m := stderrScanner.Text()
				log.Warning(m)
			}
		}()
		go func() {
			for stdoutScanner.Scan() {
				m := stdoutScanner.Text()
				log.Info(m)
			}
		}()

		waitChan := make(chan error, 10)
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(signals)

		go func() {
			waitChan <- cmd.Wait()
		}()

		select {
		case err := <-waitChan:
			if err != nil {
				if err != nil {
					log.WithError(err).Error("could not run dhclient on interface %s", ifaceName)
					return err
				}
			}
		case <-signals:
			logrus.Error("cancelled by user, killing dhclient")
			err = cmd.Process.Kill()
			if err != nil {
				log.WithError(err).Error("could not kill dhclient")
			}
			return fmt.Errorf("cancelled by user")
		}

		// remove default routes added
		routes, err := nlHandle.RouteList(nil, netlink.FAMILY_V4)
		if err != nil {
			return err
		}

		link, _ := netlink.LinkByName(ifaceName)
		if link == nil {
			log.Errorf("veth interface does not exist: %s", ifaceName)
			return fmt.Errorf("no such interface %s", ifaceName)
		}

		for _, route := range routes {
			if route.Dst == nil && route.LinkIndex == link.Attrs().Index {
				log.Infof("removing default route added by dhclient for %s", ifaceName)
				err := nlHandle.RouteDel(&route)
				if err != nil {
					log.WithError(err).Errorf("could not delete default route set on %s", ifaceName)
					return err
				}
			}
		}
	}

	return nil
}
