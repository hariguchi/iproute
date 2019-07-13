/*
   Usage: example <add | delete>


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
	usageStr = "Usage: example <add | delete>"
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
		if veth.IsNotFound(err) {
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

	if err := iproute.IfSetNS(veth.Name, ns1); err != nil {
		deleteNetwork()
		errExit(fmt.Sprintf(
			"%sIfSetNs(%s, %s): %v", banner, veth.Name, ns1, err))
	}
	if err := iproute.IfSetNS(veth.Peer, ns2); err != nil {
		deleteNetwork()
		errExit(fmt.Sprintf(
			"%sIfSetNs(%s, %s): %v", banner, veth.Peer, ns2, err))
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
