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
