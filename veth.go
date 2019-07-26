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
	"net"
)

type Veth struct {
	Link netlink.Link
	Peer netlink.Link
}

// VethGetLinkByName returns a pointer to netlink.Veth whose name is `name'
// in: name Name of veth interface
// return: 1. Pointer to netlink.Veth associated with `name'
//            undetermined otherwise.
//         2. nil if there is a veth interface whose name is `name'
//            non-nil otherwise
func VethGetLinkByName(name string) (*netlink.Veth, error) {
	if l, err := netlink.LinkByName(name); err == nil {
		switch l := l.(type) {
		case *netlink.Veth:
			return l, nil
		default:
			return nil, fmt.Errorf("VethGetLinkByName(%s): not veth", name)
		}
	} else {
		return nil, fmt.Errorf("VethGetLinkByName(%s): %v", name, err)
	}
}

// VethGetPeerLinkByName returns a pointer to netlink.Veth that is
// the peer of veth interface whose name is `name'
// in: name Name of veth interface
// return: 1. Pointer to netlink.Veth associated with the peer of `name'
//            undetermined otherwise.
//         2. nil if there is the peer of veth interface whose name is `name'
//            non-nil otherwise
func VethGetPeerLinkByName(name string) (*netlink.Veth, error) {
	if l, err := netlink.LinkByName(name); err == nil {
		switch l := l.(type) {
		case *netlink.Veth:
			if idx, err := netlink.VethPeerIndex(l); err == nil {
				if l, err := netlink.LinkByIndex(idx); err == nil {
					return l.(*netlink.Veth), nil
				} else {
					return nil, fmt.Errorf("LinkByIndex(%s): %v", name, err)
				}
			} else {
				return nil, fmt.Errorf("VethPeerIndex(%s): %v", name, err)
			}
		default:
			return nil,
				fmt.Errorf("VethGetPeerLinkByName(): %s is not veth", name)
		}
	} else {
		return nil, fmt.Errorf("VethGetPeerLinkByName(%s): %v", name, err)
	}
}

// VethGetByName returns a pointer to Veth whose name is `name'
// in: name Name of veth interface
// return: 1. Pointer to Veth associated with `name'
//            undetermined otherwise. `Veth.Peer' is nil
//            in the case the peer belongs to a different network namespace
//         2. nil if there is a veth interface whose name is `name'
//            non-nil otherwise
func VethGetByName(name string) (*Veth, error) {
	var veth Veth

	if l, err := VethGetLinkByName(name); err == nil {
		veth.Link = l
	} else {
		return nil, fmt.Errorf("VethGetLinkByName(%s) %v", name, err)
	}
	if l, err := VethGetPeerLinkByName(name); err == nil {
		veth.Peer = l
		return &veth, nil
	} else if IsNotFound(err) {
		//
		// the peer belongs to a different namespace
		//
		return &veth, nil
	} else {
		return nil, fmt.Errorf("VethGetPeerLinkByName(%s): %v", name, err)
	}
}

// VethAdd adds a veth pair. The values of all fields except
// interface names are default
// in: name Name of veth
//     peer Name of veth peer
// return 1. Pointer to Veth if success
//           nil otherwise
//        2. nil if success
//           non-nil otherwise
func VethAdd(name, peer string, up bool) (*Veth, error) {
	var (
		veth Veth
		msg  string
	)
	l := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name:        name,
			TxQLen:      DefaultTxQlen,
			MTU:         DefaultMTU,
			NumTxQueues: DefaultTxQueues,
			NumRxQueues: DefaultRxQueues,
		},
		PeerName: peer,
	}
	err := netlink.LinkAdd(l)
	if err != nil {
		return nil, err
	}
	if l, err := VethGetLinkByName(name); err == nil {
		veth.Link = l
	} else {
		return nil, fmt.Errorf("VethGetLinkByName(%s): %v", name, err)
	}
	if l, err := VethGetPeerLinkByName(name); err == nil {
		veth.Peer = l
	} else {
		return nil, fmt.Errorf("VethGetPeerLinkByName(%s): %v", name, err)
	}
	if up {
		if err := netlink.LinkSetUp(veth.Link); err != nil {
			msg = fmt.Sprintf("LinkSetUp(%s): %v", veth.Name(), err)
		}
		if err := netlink.LinkSetUp(veth.Peer); err != nil {
			if msg != "" {
				msg += ", "
			}
			msg = fmt.Sprintf("LinkSetUp(%s): %v", veth.PeerName(), err)
		}
	}
	if msg == "" {
		return &veth, nil
	}
	return &veth, fmt.Errorf(msg)
}

// VethDelete deletes the specified veth pair
// in: name Name of veth interface
// return: nil if success
//         non-nil otherwise
func VethDelete(name string) error {
	return LinkDel(name)
}

func VethIfExists(name string) (bool, error) {
	return ifExists(name, &netlink.Veth{})
}

// SetNS bind veth `v' to network namespace `nsName'
// in: nsName Name of the network namespace to bind `v'
// return nil if success
//        non-nil otherwise
func (v *Veth) SetNS(nsName string, up bool) error {
	return IfSetNS(v.Name(), nsName)
}

// SetNSbyPid bind veth `v' to network namespace whose process ID is `pid'
// in: nsName Name of the network namespace to bind `v'
// return nil if success
//        non-nil otherwise
func (v *Veth) SetNSbyPid(pid int) error {
	return IfSetNSbyPid(v.Name(), pid)
}

// UnsetNS unbinds veth `v' from network namespace whose
// process ID is `pid'
// in: nsName Name of the network namespace to unbind `v'
// return nil if success
//        non-nil otherwise
func (v *Veth) UnsetNS(nsName string) error {
	return IfUnsetNS(v.Name(), nsName)
}

// IpAddrAdd adds an IP prefix to either this or peer interface.
// in: intf Add `addr' to this interface if true
//          Add `addr' to the peer interface if false
//     addr IP prefix (IPv4 or IPv6)
//     up Bring up the interface if true
//        Do nothing otherwise
// return: nil if success
//         non-nil otherwise
func (v *Veth) IpAddrAdd(intf bool, addr *net.IPNet, up bool) error {
	var l netlink.Link

	if intf == Self {
		l = v.Link
	} else if v.Peer != nil {
		l = v.Peer
	} else {
		fs := "IpAddrAdd(%s): peer belongs to a different namespace"
		return fmt.Errorf(fs, v.Name())
	}
	if err := netlink.AddrAdd(l, &netlink.Addr{IPNet: addr}); err != nil {
		return err
	}
	if up {
		return netlink.LinkSetUp(l)
	}
	return nil
}

// IpAddrReplace replaces (or adds unless present) an IP prefix to
// either this or peer interface.
// in: intf Add `addr' to this interface if true
//          Add `addr' to the peer interface if false
//     addr IP prefix (IPv4 or IPv6)
//     up Bring up the interface if true
//        Do nothing otherwise
// return: nil if success
//         non-nil otherwise
func (v *Veth) IpAddrReplace(intf bool, addr *net.IPNet, up bool) error {
	var l netlink.Link

	if intf == Self {
		l = v.Link
	} else if v.Peer != nil {
		l = v.Peer
	} else {
		fs := "IpAddrReplace(%s): peer belongs to a different namespace"
		return fmt.Errorf(fs, v.Name())
	}
	if err := netlink.AddrReplace(l, &netlink.Addr{IPNet: addr}); err != nil {
		return err
	}
	if up {
		return netlink.LinkSetUp(l)
	}
	return nil
}

// IpAddrAdd deletes the IP prefix from either this or peer interface.
// in: intf Delete `addr' from this interface if true
//          Delete `addr' from the peer interface if false
//     addr IP prefix (IPv4 or IPv6)
// return: nil if success
//         non-nil otherwise
func (v *Veth) IpAddrDelete(intf bool, addr *net.IPNet) error {
	var l netlink.Link

	if intf == Self {
		l = v.Link
	} else if v.Peer != nil {
		l = v.Peer
	} else {
		fs := "IpAddrDelete(%s): peer belongs to a different namespace"
		return fmt.Errorf(fs, v.Name())
	}
	return netlink.AddrDel(l, &netlink.Addr{IPNet: addr})
}

func (v *Veth) Name() string {
	return v.Link.Attrs().Name
}

func (v *Veth) PeerIndex() (int, error) {
	return netlink.VethPeerIndex(v.Link.(*netlink.Veth))
}

func (v *Veth) PeerName() string {
	if v.Peer == nil {
		return ""
	}
	return v.Peer.Attrs().Name
}

func (v *Veth) TxQlen() int {
	return v.Link.Attrs().TxQLen
}

func (v *Veth) PeerTxQlen() int {
	if v.Peer == nil {
		return -1
	}
	return v.Peer.Attrs().TxQLen
}

func (v *Veth) MTU() int {
	return v.Link.Attrs().MTU
}

func (v *Veth) PeerMTU() int {
	if v.Peer == nil {
		return -1
	}
	return v.Peer.Attrs().MTU
}

func (v *Veth) NtxQs() int {
	return v.Link.Attrs().NumTxQueues
}

func (v *Veth) PeerNtxQs() int {
	if v.Peer == nil {
		return -1
	}
	return v.Peer.Attrs().NumTxQueues
}

func (v *Veth) NrxQs() int {
	if v.Peer == nil {
		return -1
	}
	return v.Link.Attrs().NumRxQueues
}

func (v *Veth) PeerNrxQs() int {
	if v.Peer == nil {
		return -1
	}
	return v.Peer.Attrs().NumRxQueues
}
