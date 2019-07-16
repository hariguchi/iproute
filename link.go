/*
Copyright 2019 Yoichi Hariguchi

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package iproute

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netlink/nl"
	"net"
)

// LinkDel deletes the specified link device (interface.)
// in: name Name of the link device (interface) to be removed
// return: nil if success
//         non-nil otherwise
func LinkDel(name string) error {
	if l, err := netlink.LinkByName(name); err == nil {
		return netlink.LinkDel(l)
	} else {
		return err
	}
}

// IfIndex returns the ifindex associated with interface `name'
// in: name Interface name
// return: 1. Ifindex (>0) of interface `name' if success
//            -1 otherwise
//         2. nil if success
//            non-nil otherwise
func IfIndex(name string) (int, error) {
	if l, err := netlink.LinkByName(name); err == nil {
		return l.Attrs().Index, nil
	} else {
		return -1, err
	}
}

// IfName retuns the interface name whose ifindex is `ifIndex'
// in: ifIndex Ifindex for the interface
// return: 1. Interface name whose ifindex is `ifIndex' if success
//            Empty string otherwise
//         2. nil if success
//            non-nil otherwise
func IfName(ifIndex int) (string, error) {
	if l, err := netlink.LinkByIndex(ifIndex); err == nil {
		return l.Attrs().Name, nil
	} else {
		return "", err
	}
}

// LinkByIndex returns Link instance whose ifindex is `ifIndex'
// in: ifIndex Ifindex for the interface
// return: 1. Link instance for the instance whose if success
//            Undetermined Link instance otherwise
func LinkByIndex(ifIndex int) (Link, error) {
	return netlink.LinkByIndex(ifIndex)
}

// LinkByIndex returns Link instance whose name is `name'
// in: ifIndex Ifindex for the interface
// return: 1. Link instance for the instance whose if success
//            Undetermined Link instance otherwise
func LinkByName(name string) (Link, error) {
	return netlink.LinkByName(name)
}

// IfUpByName brings up the specified interface
// in: name Interface name
// return: nil if success
//         non-nil otherwise
func IfUpByName(name string) error {
	if l, err := netlink.LinkByName(name); err == nil {
		return netlink.LinkSetUp(l)
	} else {
		return err
	}
}

// IfUpByName brings down the specified interface
// in: name Interface name
// return: nil if success
//         non-nil otherwise
func IfDownByName(name string) error {
	if l, err := netlink.LinkByName(name); err == nil {
		return netlink.LinkSetDown(l)
	} else {
		return err
	}
}

// IfIsUpByName returns true if the specified interface is up
// in: name Interface name
// return: 1. true if interface status is up
//            false otherwise
//         2. nil if success
//            non-nil otherwise
func IfIsUpByName(name string) (bool, error) {
	if l, err := netlink.LinkByName(name); err == nil {
		if l.Attrs().Flags&net.FlagUp != 0 {
			return true, nil
		} else {
			return false, nil
		}
	} else {
		return false, err
	}
}

// IfRename changes the name of the specified interface.
// This function brings down the interface momentarily
// in: oldName original name of the interface
//     newName new name of the interface
// return: nil if success
//         non-nil otherwise
func IfRename(oldName, newName string) error {
	errMsg := fmt.Sprintf("Error: IfRename(%s, %s): ", oldName, newName)

	if l, err := netlink.LinkByName(oldName); err == nil {
		var linkUp bool
		if l.Attrs().Flags&net.FlagUp == 0 {
			linkUp = false
		} else {
			//
			// link is up. bring it down first.
			//
			linkUp = true
			if err := netlink.LinkSetDown(l); err != nil {
				return fmt.Errorf("IfRename(%s): LinkSetName(%s): %v",
					oldName, newName, err)
			}
		}
		//
		// link is down now. Rename the interface
		//
		if err := netlink.LinkSetName(l, newName); err == nil {
			if linkUp {
				//
				// Bring up the link again
				//
				if err := netlink.LinkSetUp(l); err != nil {
					return fmt.Errorf(errMsg+"LinkSetUp(): %v", err)
				}
			}
			return nil
		} else {
			if linkUp {
				//
				// Bring up the link again
				//
				if err := netlink.LinkSetUp(l); err != nil {
					return fmt.Errorf(errMsg+"LinkSetUp(): %v", err)
				}
			}
			return fmt.Errorf(errMsg+"LinkSetName(): %v", err)
		}
	} else {
		return fmt.Errorf(errMsg+"LinkByName(): %v", err)
	}
}

// IfUnbind unbinds an interface from the master device
// in: ifName Name of interface to be unbound
// return: nil if success
//         non-nil otherwise
func IfUnbind(ifName string) error {
	if l, err := netlink.LinkByName(ifName); err == nil {
		return netlink.LinkSetNoMaster(l)
	} else {
		return err
	}
}

// IsTunnelByIndex returns true if theh specified interface is a
// tunnel interface.
// in: ifindex Ifindex of the interface to test
// return: 1. true if the interface is a tunnel interface
//            false otherwise
//         2. nil if success
//            non-nil otherwise
func IsTunnelByIndex(ifindex int) (bool, error) {
	if link, err := netlink.LinkByIndex(ifindex); err == nil {
		if link.Type() == "tun" {
			return true, nil
		} else {
			return false, nil
		}
	} else {
		return false, err
	}
}

// IsTunnelByName returns true if theh specified interface is a
// tunnel interface.
// in: name Name of the interface to test
// return: 1. true if the interface is a tunnel interface
//            false otherwise
//         2. nil if success
//            non-nil otherwise
func IsTunnelByName(name string) (bool, error) {
	if link, err := netlink.LinkByName(name); err == nil {
		if link.Type() == "tun" {
			return true, nil
		} else {
			return false, nil
		}
	} else {
		return false, err
	}
}

// IpAddrAdd adds an IP prefix to an interface
// in: name Interface name
//     addr IP prefix (IPv4 or IPv6)
//     up Bring up `name' if true
//        Do nothing otherwise
// return: nil if success
//         non-nil otherwise
func IpAddrAdd(name string, addr *net.IPNet, up bool) error {
	if l, err := netlink.LinkByName(name); err == nil {
		if err := netlink.AddrAdd(l, &netlink.Addr{IPNet: addr}); err != nil {
			return err
		}
		if up {
			return netlink.LinkSetUp(l)
		}
		return nil
	} else {
		return err
	}
}

// IpAddrDelete deletes an IP prefix from an interface
// in: `name': interface name
//     `addr': IP prefix (IPv4 or IPv6)
// return: nil if success
//         non-nil otherwise
func IpAddrDelete(name string, addr *net.IPNet) error {
	if l, err := netlink.LinkByName(name); err == nil {
		return netlink.AddrDel(l, &netlink.Addr{IPNet: addr})
	} else {
		return err
	}
}

// IpAddrReplace replaces (or adds unless present) an IP prefix
// on an interface
// in: name Interface name
//     addr IP prefix (IPv4 or IPv6)
//     up Bring up `name' if true
//        Do nothing otherwise
// return: nil if success
//         non-nil otherwise
func IpAddrReplace(name string, addr *net.IPNet, up bool) error {
	if l, err := netlink.LinkByName(name); err == nil {
		if err :=
			netlink.AddrReplace(l, &netlink.Addr{IPNet: addr}); err != nil {
			return err
		}
		if up {
			return netlink.LinkSetUp(l)
		}
		return nil
	} else {
		return err
	}
}

// IpAddrList returns a list of IP prefixes associated with the interface
// in: `name': interface name
//     `family': FAMILY_ALL, FAMILY_V4, FAMILY_V6, or FAMILY_MPLS
// return: 1. slice of *net.IPNet if success
//            nil otherwise
//         2. nil if success
//            non-nil otherwise
func IpAddrList(name string, family int) ([]*net.IPNet, error) {
	var rc []*net.IPNet

	if l, err := netlink.LinkByName(name); err == nil {
		if addr, err := netlink.AddrList(l, family); err == nil {
			for _, a := range addr {
				rc = append(rc, a.IPNet)
			}
			return rc, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

// Ipv4AddrList returns a list of IPv4 prefixes associated with the interface
// in: `name': interface name
// return: 1. slice of *net.IPNet if success
//            nil otherwise
//         2. nil if success
//            non-nil otherwise
func IPv4AddrList(name string) ([]*net.IPNet, error) {
	return IpAddrList(name, nl.FAMILY_V4)
}

// Ipv6AddrList returns a list of IPv6 prefixes associated with the interface
// in: `name': interface name
// return: 1. slice of *net.IPNet if success
//            nil otherwise
//         2. nil if success
//            non-nil otherwise
func IPv6AddrList(name string) ([]*net.IPNet, error) {
	return IpAddrList(name, nl.FAMILY_V6)
}

// IsIfPrefix returns true if `ifPrefix' belongs to interface `ifName'
// in: ifName Interface name
//     ifPrefix Pointer to net.IPNet representing the prefix to be tested
// return: 1. true if `ifPrefix' belongs to `ifName'
//            false otherwise
//         2. nil if success
//            non-nil otherwise
func IsIfPrefix(ifName string, ifPrefix *net.IPNet) (bool, error) {
	if prefixes, err := IpAddrList(ifName, nl.FAMILY_ALL); err == nil {
		for _, pfx := range prefixes {
			if IPNetEqual(pfx, ifPrefix) == true {
				return true, nil
			}
		}
		return false, err
	} else {
		return false, err
	}
}

// IsIfPrefixByName returns true if `ifPrefix' belongs to interface `ifName'
// in: ifName Interface name
//     ifPrefix IP prefix to be tested
// return: 1. true if `ifPrefix' belongs to `ifName'
//            false otherwise
//         2. nil if success
//            non-nil otherwise
func IsIfPrefixByName(ifName, ifPrefix string) (bool, error) {
	if _, addr, err := net.ParseCIDR(ifPrefix); err == nil {
		return IsIfPrefix(ifName, addr)
	} else {
		return false, err
	}
}

func IfList() ([]string, error) {
	var ifs []string
	ll, err := netlink.LinkList()
	if err != nil {
		return nil, fmt.Errorf("LinkList(): %v", err)
	}
	for _, l := range ll {
		ifs = append(ifs, l.Attrs().Name)
	}
	return ifs, nil
}
