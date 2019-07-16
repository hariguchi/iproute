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
)

type Vlan struct {
	Link *netlink.Vlan
}

// VlanAdd adds a VLAN interface to the master interface
// in: ifName Name of the master interface
//     vlanId VLAN ID
// return: 1. Name of the VLAN interface if success
//            Empty string otherwise
//         2. nil if success
//            non-nil otherwise
func VlanAdd(ifName string, vlanId uint16) (*Vlan, error) {
	if l, err := netlink.LinkByName(ifName); err == nil {
		ifName := fmt.Sprintf("%s.%d", ifName, vlanId)
		if err := netlink.LinkAdd(&netlink.Vlan{
			netlink.LinkAttrs{
				Name:        ifName,
				ParentIndex: l.Attrs().Index,
			},
			int(vlanId)}); err != nil {
			return nil, fmt.Errorf("LinkAdd(%s): %v", ifName, err)
		}
		if l, err := netlink.LinkByName(ifName); err == nil {
			switch l := l.(type) {
			case *netlink.Vlan:
				return &Vlan{Link: l}, nil
			default:
				return nil, fmt.Errorf("VlanAdd(%s): not a VLAN", l.Attrs().Name)
			}
		} else {
			return nil, fmt.Errorf("LinkByNamme(%s): %v", ifName, err)
		}
	} else {
		return nil, fmt.Errorf("LinkByName(%s): %v", ifName, err)
	}
}

// VlanDelete deletes the specified VLAN interface
// in: name Name of the VLAN interface to be deleted
// return: nil if success
//         non-nil otherwise
func VlanDelete(name string) error {
	return LinkDel(name)
}

func (vlan *Vlan) Name() string {
	return vlan.Link.Attrs().Name
}

func (vlan *Vlan) VlanId() int {
	return vlan.Link.VlanId
}
