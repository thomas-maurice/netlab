# netlab

`netlab` is a helper to setup/teardown my network lab. I regularly need to create bridges, veth interfaces and whatnot to experiment stuff with OpenWRT like demonstrated [here](https://blog.thomas.maurice.fr/posts/virtualise-openwrt/). For this I need `virt-manager` to use  a bunch of bridge interfaces to connect the port of the router too, and since I use this to configure my home router as well, I need to be able to interact with VLAN interfaces and what not, which is annoying. I wrote a small bash script available [here](https://gist.github.com/thomas-maurice/b07dca00695f952775a5e99fe2aad8f7) but it became complicated and not really manageable for VLANs and such. So I wrote this.

## How does it work ?

You can do basically 4 things:
* Configure the interfaces
* Destroy the interfaces
* Start `dhclient` to get a DHCP address form the router
* Stop the `dhclient` instances that were running

:warning: `dhclient` adds a default route, you might not want that, so it gets removed by the binary

For example, running `sudo ./netlab up -c network.yaml` would create the following interfaces:
```
23: sw-r0-eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UNKNOWN mode DEFAULT group default 
    link/ether 1a:90:be:4a:b6:79 brd ff:ff:ff:ff:ff:ff
24: sw-r0-eth1: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP mode DEFAULT group default 
    link/ether 76:d9:4a:24:7e:7b brd ff:ff:ff:ff:ff:ff
25: sw-r0-eth2: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UNKNOWN mode DEFAULT group default 
    link/ether d6:a8:34:8c:63:86 brd ff:ff:ff:ff:ff:ff
26: sw-r0-eth3: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UNKNOWN mode DEFAULT group default 
    link/ether 02:56:5f:6b:0d:90 brd ff:ff:ff:ff:ff:ff
27: sw-r0-eth4: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UNKNOWN mode DEFAULT group default 
    link/ether 72:01:0c:9d:23:d5 brd ff:ff:ff:ff:ff:ff
28: sw-r0-eth5: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP mode DEFAULT group default 
    link/ether 6a:ab:d5:b1:ed:b2 brd ff:ff:ff:ff:ff:ff
29: net1-1@net1-0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue master sw-r0-eth5 state UP mode DEFAULT group default 
    link/ether 6a:ab:d5:b1:ed:b2 brd ff:ff:ff:ff:ff:ff
30: net1-0@net1-1: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP mode DEFAULT group default 
    link/ether 26:b0:5b:a1:a3:f4 brd ff:ff:ff:ff:ff:ff
31: net0-1@net0-0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue master sw-r0-eth1 state UP mode DEFAULT group default 
    link/ether 76:d9:4a:24:7e:7b brd ff:ff:ff:ff:ff:ff
32: net0-0@net0-1: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP mode DEFAULT group default 
    link/ether f2:31:75:f6:d1:23 brd ff:ff:ff:ff:ff:ff
33: net0-0.16@net0-0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP mode DEFAULT group default 
    link/ether f2:31:75:f6:d1:23 brd ff:ff:ff:ff:ff:ff
34: net0-0.11@net0-0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP mode DEFAULT group default 
    link/ether f2:31:75:f6:d1:23 brd ff:ff:ff:ff:ff:ff
35: net0-0.12@net0-0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP mode DEFAULT group default 
    link/ether f2:31:75:f6:d1:23 brd ff:ff:ff:ff:ff:ff
36: net0-0.13@net0-0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP mode DEFAULT group default 
    link/ether f2:31:75:f6:d1:23 brd ff:ff:ff:ff:ff:ff
37: net0-0.14@net0-0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP mode DEFAULT group default 
    link/ether f2:31:75:f6:d1:23 brd ff:ff:ff:ff:ff:ff
```

You can clean up running `sudo ./netlab down`. You can also request DHCP addresses after starting the vm like so `sudo ./netlab dhcp-up` and kill the `dhclient` instances with a `dhcp-down`, simple as that.

## Configuration
The configuration file looks like this and is self explanatory
```yaml
bridges:
  # These are the interfaces we are plugging the router's VM on on virt-manager
  sw-r0-eth0:
  sw-r0-eth1:
  sw-r0-eth2:
  sw-r0-eth3:
  sw-r0-eth4:
  sw-r0-eth5:
veths:
  # this is a "link" plugged onto the last router's interface
  net1:
    dhcp: true
    master: sw-r0-eth5
  # This is an interface plugged on the eth1 of the switch
  # and is configured to have several VLANs on it. This mimicks
  # my home setup, this would be the port connected to my fat switch
  net0:
    dhcp: true
    master: sw-r0-eth1
    vlans:
      11:
          dhcp: true
      12:
          dhcp: false
      13:
          dhcp: true
      14:
          dhcp: false
      16:
          dhcp: false
```