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
)

type Vrf struct {
	Link *netlink.Vrf
}

// VrfGetLinkByIndex returns a pointer to Vrf whose ifindex is `idx'
// in: idx Ifindex of the target VRF
// return: 1. Pointer to netlink.Vrf whose ifindex is `idx' if success
//            nil otherwise
//         2. nil if success
//            non-nil otherwise
func VrfGetByIndex(idx int) (*Vrf, error) {
	if l, err := netlink.LinkByIndex(idx); err == nil {
		switch l := l.(type) {
		case *netlink.Vrf:
			return &Vrf{Link: l}, nil
		default:
			return nil, fmt.Errorf("VrfGetByIndex(%d): not a VRF", idx)
		}
	} else {
		return nil, err
	}
}

// VrfGetLinkByName returns a pointer to Vrf whose name is `name'
// in: name Name of VRF
// return: 1. Pointer to netlink.Vrf associated with `name'
//            undetermined otherwise.
//         2. nil if there is a VRF whose name is `name'
//            non-nil otherwise
func VrfGetByName(name string) (*Vrf, error) {
	if l, err := netlink.LinkByName(name); err == nil {
		switch l := l.(type) {
		case *netlink.Vrf:
			return &Vrf{Link: l}, nil
		default:
			return nil, fmt.Errorf("VrfGetByName(%s): not a VRF", name)
		}
	} else {
		return nil, err
	}
}

// VrfAdd adds VRF whose name is `name' and whose table id is `tid'
// in: name Name of VRF to be added
//     tid Table ID for VRF `name'
// return: nil if success
//         non-nil otherwise
func VrfAdd(name string, tid uint32, up bool) (*Vrf, error) {
	err := netlink.LinkAdd(&netlink.Vrf{
		LinkAttrs: netlink.LinkAttrs{Name: name},
		Table:     uint32(tid),
	})
	if err != nil {
		return nil, err
	}
	if vrf, err := VrfGetByName(name); err == nil {
		if up {
			vrf.IfUp()
		}
		return vrf, nil
	} else {
		return nil, err
	}
}

// VrfDelete deletes VRF whose name is `name'
// in: name Name of VRF
// return: nil if success
//         non-nil otherwise
func VrfDelete(name string) error {
	return LinkDel(name)
}

// VrfBindIf binds an interface to a VRF
// in: vrfName Name of VRF
//     ifName Name of interface to be bound to VRF `vrfName'
// return: nil if success
//         non-nil otherwise
func VrfBindIf(vrfName, ifName string) error {
	if vrf, err := VrfGetByName(vrfName); err == nil {
		return vrf.BindIf(ifName)
	} else {
		return fmt.Errorf("VrfGetByName(%s): %v", vrfName, err)
	}
}

// VrfGetRoutesByTid returns a slice of netlink.Route whose
// table id is `tid', family is `family', and table type is `tableType'
// in: tid Table ID
//     family FAMILY_ALL, FAMILY_V4, FAMILY_V6, or FAMILY_MPLS
//     tableType RTN_UNSPEC, RTN_UNICAST, RTN_LOCAL, RTN_BROADCAST,
//               RTN_ANYCAST,RTN_MULTICAST, RTN_BLACKHOLE, RTN_UNREACHABLE,
//               RTN_PROHIBIT, RTN_THROW, RTN_NAT, or RTN_XRESOLVE
// return: 1. slice of netlink.Route if success
//            undetermined otherwise
//         2. nil if success
//            non-nil otherwise
func VrfGetRoutesByTid(tid int, family int, tableType int) (Routes, error) {
	routeFilter := &netlink.Route{
		Table: tid,
		Type:  tableType,
	}
	filterMask := RT_FILTER_TABLE | RT_FILTER_TYPE
	return netlink.RouteListFiltered(family, routeFilter, filterMask)
}

// VrfGetRoutesByName returns a slice of netlink.Route belonging to the VRF
// whose nanme is `name', family is `family', and table type is `tableType'
// in: name Name of VRF
//     family FAMILY_ALL, FAMILY_V4, FAMILY_V6, or FAMILY_MPLS
//     tableType RTN_UNSPEC, RTN_UNICAST, RTN_LOCAL, RTN_BROADCAST,
//               RTN_ANYCAST,RTN_MULTICAST, RTN_BLACKHOLE, RTN_UNREACHABLE,
//               RTN_PROHIBIT, RTN_THROW, RTN_NAT, or RTN_XRESOLVE
// return: 1. slice of netlink.Route if success
//            undetermined otherwise
//         2. nil if success
//            non-nil otherwise
func VrfGetRoutesByName(name string, family int, tblType int) (Routes, error) {
	if vrf, err := VrfGetByName(name); err == nil {
		return VrfGetRoutesByTid(int(vrf.Tid()), family, tblType)
	} else {
		errMsg := fmt.Sprintf("VrfGetRoutesByName(%s): ", vrf.Name())
		return nil, fmt.Errorf(errMsg+"%v", err)
	}
}

// VrfGetIPv4routesByName returns a slice of IPv4 netlink.Route
// belonging to the VRF whose nanme is `name'
// in: name Name of VRF
// return: 1. slice of netlink.Route if success. All of them are IPv4
//            undetermined otherwise
//         2. nil if success
//            non-nil otherwise
func VrfGetIPv4routesByName(name string) (Routes, error) {
	if vrf, err := VrfGetByName(name); err == nil {
		return VrfGetRoutesByTid(int(vrf.Tid()), nl.FAMILY_V4, RTN_UNICAST)
	} else {
		return Routes{}, err
	}
}

// VrfGetIPv4localRoutesByName returns a slice of IPv4 netlink.Route of
// local routes belonging to the VRF whose nanme is `name'
// in: name Name of VRF
// return: 1. slice of netlink.Route if success. All of them are IPv4
//            undetermined otherwise
//         2. nil if success
//            non-nil otherwise
func VrfGetIPv4localRoutes(vrf string) (Routes, error) {
	return VrfGetRoutesByName(vrf, nl.FAMILY_V4, RTN_LOCAL)
}

// VrfGetIPv6routesByName returns a slice of IPv6 netlink.Route
// belonging to the VRF whose nanme is `name'
// in: name Name of VRF
// return: 1. slice of netlink.Route if success. All of them are IPv4
//            undetermined otherwise
//         2. nil if success
//            non-nil otherwise
func VrfGetIPv6routesByName(name string) (Routes, error) {
	if vrf, err := VrfGetByName(name); err == nil {
		return VrfGetRoutesByTid(int(vrf.Tid()), nl.FAMILY_V6, RTN_UNICAST)
	} else {
		return Routes{}, err
	}
}

// VrfAddRouteByName adds a route to a VRF
// in: name Name of the target VRF
//     r Pointer to the route to be added
// return: nil if success
//         non-nil otherwise
func VrfAddRouteByName(name string, r *Route) error {
	errMsg := fmt.Sprintf("VrfAddRouteByName(%s, %v): ", name, r)
	if vrf, err := VrfGetByName(name); err == nil {
		r.Table = int(vrf.Tid())
		return netlink.RouteAdd(r)
	} else {
		return fmt.Errorf(errMsg+"VrfGetByName(): %v", err)
	}
}

// VrfDeleteRouteByName deletes a route in a VRF.
// in: name Name of the target VRF
//     r Pointer to the route to be deleted
// return: nil if success
//         non-nil otherwise
func VrfDeleteRouteByName(name string, r *Route) error {
	errMsg := fmt.Sprintf("VrfDeleteRouteByName(%s, %v): ", name, r)
	if vrf, err := VrfGetByName(name); err == nil {
		r.Table = int(vrf.Tid())
		return netlink.RouteDel(r)
	} else {
		return fmt.Errorf(errMsg+"VrfGetByName(): %v", err)
	}
}

// VrfReplaceRouteByName replaces the existing route in a VRF.
// The route is added to the VRF unless it exists.
// in: name Name of the target VRF
//     r Pointer to the route to be added
// return: nil if success
//         non-nil otherwise
func VrfReplaceRouteByName(name string, r *Route) error {
	errMsg := fmt.Sprintf("VrfReplaceRouteByName(%s, %v): ", name, r)
	if vrf, err := VrfGetByName(name); err == nil {
		r.Table = int(vrf.Tid())
		return netlink.RouteReplace(r)
	} else {
		return fmt.Errorf(errMsg+"VrfGetByName(): %v", err)
	}
}

// Equal checks if two VRFs are identical or not.
// in: other Pointer to `Vrf'
// return: true if VRF `other' has the same name, ifindex, and table id
//         false otherwise.
func (vrf *Vrf) Equal(other *Vrf) bool {
	if vrf.Name() == other.Name() &&
		vrf.Index() == other.Index() && vrf.Tid() == other.Tid() {
		return true
	}
	return false
}

func (vrf *Vrf) Name() string {
	return vrf.Link.Attrs().Name
}

func (vrf *Vrf) Index() int {
	return vrf.Link.Attrs().Index
}

func (vrf *Vrf) Tid() uint32 {
	return vrf.Link.Table
}

func (vrf *Vrf) IfUp() error {
	return netlink.LinkSetUp(vrf.Link)
}

func (vrf *Vrf) IfDown() error {
	return netlink.LinkSetDown(vrf.Link)
}

// VrfBindIf binds an interface to a VRF
// in: ifName Name of interface to be bound to VRF `vrfName'
// return: nil if success
//         non-nil otherwise
func (vrf *Vrf) BindIf(ifName string) error {
	if l, err := netlink.LinkByName(ifName); err == nil {
		return netlink.LinkSetMasterByIndex(l, vrf.Index())
	} else {
		return fmt.Errorf("LinkByName(%s): %v", ifName, err)
	}

}
