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
	"regexp"
)

type Veth struct {
	Name   string
	Peer   string
	TxQlen int
	MTU    int
	NtxQs  int
	NrxQs  int
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
//            undetermined otherwise.
//         2. nil if there is a veth interface whose name is `name'
//            non-nil otherwise
func VethGetByName(name string) (*Veth, error) {
	if l, err := VethGetPeerLinkByName(name); err == nil {
		return &Veth{
			Name:   name,
			Peer:   l.Attrs().Name,
			TxQlen: l.Attrs().TxQLen,
			MTU:    l.Attrs().MTU,
			NtxQs:  l.Attrs().NumTxQueues,
			NrxQs:  l.Attrs().NumRxQueues,
		}, nil
	} else {
		return nil, fmt.Errorf("VethGetByName(%s): %v", name, err)
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
	veth := Veth{
		Name:   name,
		Peer:   peer,
		TxQlen: DefaultTxQlen,
		MTU:    DefaultMTU,
		NtxQs:  DefaultTxQueues,
		NrxQs:  DefaultRxQueues,
	}
	if err := veth.Add(up); err == nil {
		return &veth, nil
	} else {
		return nil, err
	}
}

// VethDelete deletes the specified veth pair
// in: name Name of veth interface
// return: nil if success
//         non-nil otherwise
func VethDelete(name string) error {
	return LinkDel(name)
}

// IsNotFound returns true if `err' contains "not found"
// in: err Error from VethGetByName()
// return: true if err contains "not found"
//         false otherwise
func (v *Veth) IsNotFound(err error) bool {
	re := regexp.MustCompile(`not found`)
	if re.MatchString(fmt.Sprint(err)) {
		return true
	}
	return false
}

// Add adds a veth pair. All fields in Veth must be filled.
func (v *Veth) Add(up bool) error {
	l := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name:        v.Name,
			TxQLen:      v.TxQlen,
			MTU:         v.MTU,
			NumTxQueues: v.NtxQs,
			NumRxQueues: v.NrxQs,
		},
		PeerName: v.Peer,
	}
	err := netlink.LinkAdd(l)
	if err != nil {
		return err
	}
	if up {
		return netlink.LinkSetUp(l)
	}
	return nil
}
