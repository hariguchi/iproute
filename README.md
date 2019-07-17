
# iproute - easy-to-use golang Linux Netlink Library #

The iproute package provides an easy-to-use wrapper for  
[github.com/vishvananda/netlink](https://github.com/vishvananda/netlink/).

## Examples ##

### examples/example1.go ###
Bridge and VRF
```go
/*
   Usage: example1 <add | delete>


   Add or delete the following network namespaces

                        +------+
                        | vrf1 |
                        +-+--+-+
  vrf12-br1:192.168.1.2/24|  |vrf11-br1:192.168.1.1/24
                          |  |
                 br1-vrf12|  |br1-vrf11
                        +-+--+-+
                        | br1  |
                        +------+
*/

package main

import (
	"fmt"
	"github.com/hariguchi/iproute"
	"net"
	"os"
	"runtime"
)

const (
	br1      = "br1"
	vrf1     = "vrf1"
	vrf11    = "vrf11"
	vrf12    = "vrf12"
	vrf11a   = "192.168.1.1/24"
	vrf12a   = "192.168.1.2/24"
	usageStr = "Usage: example1 <add | delete>"
)

var (
	vethBase = [2][2]string{{vrf11, br1}, {vrf12, br1}}
	ifPrefix = [2]string{vrf11a, vrf12a}
)

func usage() {
	fmt.Fprint(os.Stderr, usageStr, "\n")
	os.Exit(1)
}

func errExit(s string) {
	fmt.Fprint(os.Stderr, "ERROR: ", s, "\n")
	os.Exit(1)
}

func vethNames(if1, if2 string) (string, string) {
	return if1 + "-" + if2, if2 + "-" + if1
}

func addNetwork() {
	var (
		err  error
		br   *iproute.Bridge
		vrf  *iproute.Vrf
		veth [2]*iproute.Veth
	)
	banner := "addNetwork(): "

	//
	// Create a bridge
	//
	br, err = iproute.BridgeGetByName(br1)
	if err != nil {
		if iproute.IsNotFound(err) {
			br, err = iproute.BridgeAdd(br1, iproute.Up)
			if err != nil {
				msg := fmt.Sprintf("%sBridgeAdd(%s, up): %v", banner, br1, err)
				errExit(msg)
			}
			fmt.Fprintf(os.Stderr, "Added bridge %s\n", br.Name())
		} else {
			msg := fmt.Sprintf("%sBridgeGetByName(%s): %v",
				banner, br.Name(), err)
			errExit(msg)
		}
	} else {
		fmt.Fprintf(os.Stderr, "bridge %s already exists\n", br.Name())
	}

	//
	// Create a VRF
	//
	vrf, err = iproute.VrfGetByName(vrf1)
	if err != nil {
		if iproute.IsNotFound(err) {
			vrf, err = iproute.VrfAdd(vrf1, 1, iproute.Up)
			if err != nil {
				msg := fmt.Sprintf("%sVrfAdd(%s, up): %v",
					banner, vrf1, err)
				errExit(msg)
			}
			fmt.Fprintf(os.Stderr, "Added VRF %s\n", vrf.Name())
		} else {
			msg := fmt.Sprintf("%sVrfGetByName(%s): %v",
				banner, vrf.Name(), err)
			errExit(msg)
		}
	} else {
		fmt.Fprintf(os.Stderr, "VRF %s already exists\n", vrf.Name())
	}

	//
	// Create veth interfaces,
	// connect them to the VRF and bridge, and
	// assign IP prefixes to them
	//
	for i := 0; i < len(vethBase); i++ {
		//
		// Create a veth pair
		//
		if1, if2 := vethNames(vethBase[i][0], vethBase[i][1])
		veth[i], err = iproute.VethGetByName(if1)
		if err != nil {
			if iproute.IsNotFound(err) {
				veth[i], err = iproute.VethAdd(if1, if2, iproute.Up)
				if err != nil {
					msg := fmt.Sprintf("%sVethAdd(%s, %s): %v",
						banner, if1, if2, err)
					errExit(msg)
				}
				fmt.Fprintf(os.Stderr, "Added veth pairs: %s, %s\n",
					veth[i].Name(), veth[i].PeerName())
			} else {
				msg := fmt.Sprintf("%sVethGetByName(%s): %v", banner, if1, err)
				errExit(msg)
			}
		} else {
			fmt.Fprintf(os.Stderr, "veth pair %s, %s already exists\n",
				veth[i].Name(), veth[i].PeerName())
		}
		//
		// bind veth[i].Name() to VRF, PeerName() to bridge
		//
		if err = vrf.BindIf(veth[i].Name()); err != nil {
			msg := fmt.Sprintf("vrf.BindIf(%s): %v", veth[i].Name, err)
			errExit(msg)
		}
		if err = br.BindIf(veth[i].PeerName()); err != nil {
			msg := fmt.Sprintf("br.BindIf(%s): %v", veth[i].Name, err)
			errExit(msg)
		}
		if a, p, err := net.ParseCIDR(ifPrefix[i]); err == nil {
			p.IP = a
			if err := veth[i].IpAddrReplace(iproute.Self, p, iproute.Up); err != nil {
				if iproute.IsExist(err) {
					fmt.Fprintf(os.Stderr, "%s on %s already exists\n",
						veth[i].Name(), ifPrefix[i])
				} else {
					msg := fmt.Sprintf("IpAddrAdd(%s): %v", veth[i].Name(), err)
					errExit(msg)
				}
			}
		} else {
			msg := fmt.Sprintf("ParseCIDR(%s): %v", ifPrefix[i], err)
			errExit(msg)
		}
	}
}

func deleteNetwork() {
	var (
		err  error
		veth [2]*iproute.Veth
	)
	//banner := "deleteNetwork(): "

	//
	// Unbind veth interfaces
	//
	for i := 0; i < len(vethBase); i++ {
		if1, _ := vethNames(vethBase[i][0], vethBase[i][1])
		veth[i], err = iproute.VethGetByName(if1)
		if err == nil {
			if err := iproute.IfUnbind(veth[i].Name()); err != nil {
				fmt.Fprintf(os.Stderr, "IfUnbind(%s): %v", veth[i].Name(), err)
			}
		} else {
			if iproute.IsNotFound(err) {
				fmt.Fprintf(os.Stderr, "veth %s does not exist\n", if1)
			} else {
				fmt.Fprintf(os.Stderr, "VethGetByName(%s): %v\n", if1, err)
			}
		}
		if err = iproute.IfUnbind(veth[i].PeerName()); err != nil {
			fmt.Fprintf(os.Stderr, "IfUnbind(%s): %v\n", veth[i].PeerName(), err)
		}
		iproute.VethDelete(veth[i].Name())
	}
	//
	// delete bridge and vrf
	//
	if err := iproute.VrfDelete(vrf1); err != nil {
		fmt.Fprintf(os.Stderr, "VrfDelete(%s): %v\n", vrf1, err)
	}
	if err := iproute.BridgeDelete(br1); err != nil {
		fmt.Fprintf(os.Stderr, "BridgeDelete(%s): %v\n", br1, err)
	}

}

func main() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if len(os.Args) <= 1 {
		usage()
	}
	switch os.Args[1] {
	case "add":
		addNetwork()
	case "delete":
		deleteNetwork()
	default:
		usage()
	}
	os.Exit(0)
}
```


### examples/example2.go ###
Network Namespaces

```go
/*
   Usage: example2 <add | delete>


   Add or delete the following network namespaces

   +-----+
   | ns1 |
   +--+--+
      |ns1-ns2: 192.168.1.1/24
      |
      |ns2-ns1: 192.168.1.2/24
   +--+--+
   | ns2 |
   +-----+
*/

package main

import (
	"fmt"
	netns "github.com/hariguchi/go_netns"
	"github.com/hariguchi/iproute"
	"net"
	"os"
	"runtime"
)

const (
	ns1      = "ns1"
	ns2      = "ns2"
	veth0    = "ns1-ns2"
	veth1    = "ns2-ns1"
	usageStr = "Usage: example2 <add | delete>"
	ns1ns2a  = "192.168.1.1/24"
	ns2ns1a  = "192.168.1.2/24"
)

func usage() {
	fmt.Fprint(os.Stderr, usageStr, "\n")
	os.Exit(1)
}

func errExit(s string) {
	fmt.Fprint(os.Stderr, "ERROR: ", s, "\n")
	os.Exit(1)
}

func addNetwork() {
	var (
		d1   netns.NsDesc
		d2   netns.NsDesc
		orig netns.NsDesc
	)
	banner := "addNetwork(): "
	veth, err := iproute.VethGetByName(veth0)
	if err != nil {
		if iproute.IsNotFound(err) {
			if veth, err = iproute.VethAdd(veth0, veth1, true); err != nil {
				s := fmt.Sprintf("%saddNetwork(): %v", banner, err)
				errExit(s)
			}
		} else {
			errExit(fmt.Sprintf("VethGetByName(%s): %v", veth0, err))
		}
	}
	d1, err = netns.AddByName(ns1)
	if err != nil {
		errExit(fmt.Sprintf("%sAddByName(%s)", banner, ns1))
	}
	defer d1.Close()

	d2, err = netns.AddByName(ns2)
	if err != nil {
		errExit(fmt.Sprintf("%sAddByName(%s)", banner, ns2))
	}
	defer d2.Close()

	if err := iproute.IfSetNS(veth.Name(), ns1); err != nil {
		deleteNetwork()
		errExit(fmt.Sprintf(
			"%sIfSetNs(%s, %s): %v", banner, veth.Name(), ns1, err))
	}
	if err := iproute.IfSetNS(veth.PeerName(), ns2); err != nil {
		deleteNetwork()
		errExit(fmt.Sprintf(
			"%sIfSetNs(%s, %s): %v", banner, veth.PeerName(), ns2, err))
	}
	orig, err = netns.Get()
	if err != nil {
		deleteNetwork()
		errExit(fmt.Sprintf("%sGetMyHandle(): %v", banner, err))
	}
	defer orig.Close()

	//
	// switch namespace to d1
	//
	if err := d1.Set(); err != nil {
		deleteNetwork()
		errExit(fmt.Sprintf("%sSet(%s): %v", banner, ns1, err))
	}
	if err := iproute.IfUpByName("lo"); err != nil {
		fmt.Fprintf(os.Stderr, "%sd1: IfUpByName(lo): %v", banner, err)
	}
	if a, p, err := net.ParseCIDR(ns1ns2a); err == nil {
		p.IP = a
		if err := iproute.IpAddrAdd(veth0, p, true); err != nil {
			fmt.Fprintf(os.Stderr,
				"ERROR: IpAddrAdd(%s, %s): %v\n", veth0, p.String(), err)
		}
	} else {
		fmt.Fprintf(os.Stderr, "ERROR: ParseCIDR(%s), %v\n", ns1ns2a, err)
	}
	//
	// switch namespace to d2
	//
	if err := d2.Set(); err != nil {
		deleteNetwork()
		errExit(fmt.Sprintf("%sSet(%s): %v", banner, ns2, err))
	}
	if err := iproute.IfUpByName("lo"); err != nil {
		fmt.Fprintf(os.Stderr, "%sd2: IfUpByName(lo): %v", banner, err)
	}
	if a, p, err := net.ParseCIDR(ns2ns1a); err == nil {
		p.IP = a
		if err := iproute.IpAddrAdd(veth1, p, true); err != nil {
			fmt.Fprintf(os.Stderr,
				"ERROR: IpAddrAdd(%s, %s): %v\n", veth1, p.String(), err)
		}
	} else {
		fmt.Fprintf(os.Stderr, "ERROR: ParseCIDR(%s), %v\n", ns2ns1a, err)
	}
	//
	// switch namespace back to orig
	//
	if err := orig.Set(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: orig.Set(): %v\n", err)
	}
}

func deleteNetwork() {
	orig, err := netns.Get()
	if err != nil {
		errExit(fmt.Sprintf("Get(): %v", err))
	}
	defer orig.Close()

	if d1, err := netns.GetByName(ns1); err == nil {
		//
		// switched namespace to d1
		//
		if err := d1.Set(); err == nil {
			if err := iproute.IfUnsetNS(veth0, d1.Name); err != nil {
				fmt.Fprintf(os.Stderr,
					"ERROR: IfUnsetNS(%s, %s): %v\n", veth0, d1.Name, err)
			}
		} else {
			fmt.Fprintf(os.Stderr, "ERROR: d1.Set(): %v\n", err)
		}
		if err := d1.Delete(); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: d1.Delete(): %v\n", err)
		}
	} else {
		fmt.Fprintf(os.Stderr, "ERROR: GetByName(%s): %v\n", ns1, err)
	}
	if d2, err := netns.GetByName(ns2); err == nil {
		defer netns.DeleteByName(ns2)
		//
		// switched namespace to d2
		//
		if err := d2.Set(); err == nil {
			if err := iproute.IfUnsetNS(veth1, d2.Name); err != nil {
				fmt.Fprintf(os.Stderr,
					"ERROR: IfUnsetNS(%s, %s): %v\n", veth1, d2.Name, err)
			}
		} else {
			fmt.Fprintf(os.Stderr, "ERROR: d2.Set(): %v\n", err)
		}
		if err := d2.Delete(); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: d2.Delete(): %v\n", err)
		}
	} else {
		fmt.Fprintf(os.Stderr, "ERROR: GetByName(%s): %v\n", ns2, err)
	}
	//
	// switch namespace back to orig
	//
	if err := orig.Set(); err != nil {
		errExit(fmt.Sprintf("orig.Set(): %v", err))
	}
	if err := iproute.VethDelete(veth0); err != nil {
		errExit(fmt.Sprintf("VethDelete(%s): %v", veth0, err))
	}
}

func main() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if len(os.Args) <= 1 {
		usage()
	}
	switch os.Args[1] {
	case "add":
		addNetwork()
	case "delete":
		deleteNetwork()
	default:
		usage()
	}
	os.Exit(0)
}
```
