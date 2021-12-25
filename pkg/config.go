package netlab

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

const (
	maxVethNameLength   = 9 // cause + "-0" and "-1" and potentially a VLAN ID
	maxBridgeNameLength = 16
)

type Config struct {
	Bridges map[string]BridgeOptions `yaml:"bridges"`
	Veths   map[string]Veth          `yaml:"veths"`
}

func (c *Config) Validate() error {
	// Validate bridges
	for bridgeName := range c.Veths {
		if len(bridgeName) > maxBridgeNameLength {
			return fmt.Errorf("invalid name %s, should be below length %d (currently %d)", bridgeName, maxBridgeNameLength, len(bridgeName))
		}
	}

	// Validate Veths
	for vethName, veth := range c.Veths {
		if len(vethName) > maxVethNameLength {
			return fmt.Errorf("invalid name %s, should be below length %d (currently %d)", vethName, maxVethNameLength, len(vethName))
		}

		if _, ok := c.Bridges[veth.Master]; !ok {
			return fmt.Errorf("no such master bridge %s", veth.Master)
		}

		for vid := range veth.VLANs {
			if vid < 1 || vid > 4000 {
				return fmt.Errorf("invalid VLAN ID %d", vid)
			}
		}
	}

	return nil
}

type BridgeOptions struct {
}

type Veth struct {
	Master string       `yaml:"master"` // master bridge for the "-1" end of the pair
	DHCP   bool         `yaml:"dhcp"`   // Wheter or not enable DHCP on the interface
	VLANs  map[int]VLAN `yaml:"vlans"`  // VLANs on the interface
}

type VLAN struct {
	DHCP bool `yaml:"dhcp"` // Whether or not we should request an address via DHCP
}

func LoadConfigFromBytes(cfgBytes []byte) (*Config, error) {
	var cfg Config
	err := yaml.Unmarshal(cfgBytes, &cfg)
	return &cfg, err
}

func LoadConfigFromFile(fileName string) (*Config, error) {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	return LoadConfigFromBytes(b)
}
