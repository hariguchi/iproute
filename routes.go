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

// NewRoute creates a Route instance from the destination prefix
// and the next-hop
// in: dst Destination IP prefix
//     nh Slice of Next-hop IP addresses
// return: 1. Route instance associated with the parameters if success
//            undetermined Route instance otherwise
//         2. nil if success
//            non-nil otherwise
func NewRoute(dst *net.IPNet, nh IPs) (Route, error) {
	if dst == nil {
		return Route{}, fmt.Errorf("ERROR: NewRoute(): dst is nil")
	}
	errMsg := fmt.Sprintf("ERROR: NewRoute(%v, %v): ", dst, nh)
	if len(nh) <= 0 {
		return Route{}, fmt.Errorf(errMsg + "# of nh is <= 0")
	}
	r := Route{Dst: dst}
	for _, ipa := range nh {
		r.MultiPath = append(r.MultiPath, &NHinfo{Gw: ipa})
	}
	return r, nil
}

// AddNextHop adds a next-hop to the given route
// in: nh Next-hop IP addresses
// in,out: r Pointer to the target route
func AddNextHop(r *Route, nh *NHinfo) {
	r.MultiPath = append(r.MultiPath, nh)
}

// SetOnlink sets the onlink flag in either Route or NHinfo instance
// in: i Pointer to either Route or NHinfo instance
// return: nil if success
//         non-nil otherwise
func SetOnlink(i interface{}) error {
	var err error = nil

	switch v := i.(type) {
	case *Route:
		v.Flags |= int(FLAG_ONLINK)
	case *NHinfo:
		v.Flags |= int(FLAG_ONLINK)
	default:
		err = fmt.Errorf("ERROR: ClearOnlink(%v): wrong type", v)
	}
	return err
}

// ClearOnlink unsets the onlink flag in either Route or NHinfo instance
// in: i Pointer to either Route or NHinfo instance
// return: nil if success
//         non-nil otherwise
func ClearOnlink(i interface{}) error {
	var err error = nil

	switch v := i.(type) {
	case *Route:
		v.Flags &^= int(FLAG_ONLINK)
	case *NHinfo:
		v.Flags &^= int(FLAG_ONLINK)
	default:
		err = fmt.Errorf("ERROR: ClearOnlink(%v): wrong type", v)
	}
	return err
}

// SetPervasive sets the pervasive flag in either Route or NHinfo instance
// in: i Pointer to either Route or NHinfo instance
// return: nil if success
//         non-nil otherwise
func SetPervasive(i interface{}) error {
	var err error = nil

	switch v := i.(type) {
	case *Route:
		v.Flags |= int(FLAG_PERVASIVE)
	case *NHinfo:
		v.Flags |= int(FLAG_PERVASIVE)
	default:
		err = fmt.Errorf("ERROR: SetPervasive(%v): wrong type", v)
	}
	return err
}

// ClearPervasive unsets the pervasive flag in either Route or NHinfo instance
// in: i Pointer to either Route or NHinfo instance
// return: nil if success
//         non-nil otherwise
func ClearPervasive(i interface{}) error {
	var err error = nil

	switch v := i.(type) {
	case *Route:
		v.Flags &^= int(FLAG_PERVASIVE)
	case *NHinfo:
		v.Flags &^= int(FLAG_PERVASIVE)
	default:
		err = fmt.Errorf("ERROR: ClearPervasive(%v): wrong type", v)
	}
	return err
}

// GetRoutes returns a slice of netlink.Route
// whose family is `family', and table type is `tableType'
// in: family FAMILY_ALL, FAMILY_V4, FAMILY_V6, or FAMILY_MPLS
//     tableType RTN_UNSPEC, RTN_UNICAST, RTN_LOCAL, RTN_BROADCAST,
//               RTN_ANYCAST,RTN_MULTICAST, RTN_BLACKHOLE, RTN_UNREACHABLE,
//               RTN_PROHIBIT, RTN_THROW, RTN_NAT, or RTN_XRESOLVE
// return: 1. slice of netlink.Route if success
//            undetermined otherwise
//         2. nil if success
//            non-nil otherwise
func GetRoutes(family int, tblType int) (Routes, error) {
	return VrfGetRoutesByTid(0, family, tblType)
}

// GetIPv4routes returns a slice of IPv4 netlink.Route
// return: 1. slice of netlink.Route if success. All of them are IPv4
//            undetermined otherwise
//         2. nil if success
//            non-nil otherwise
func GetIPv4routes() (Routes, error) {
	return VrfGetRoutesByTid(0, nl.FAMILY_V4, RTN_UNICAST)
}

// GetIPv4localRoutes returns a slice of IPv4 netlink.Route of
// local routes
// return: 1. slice of netlink.Route if success. All of them are IPv4
//            undetermined otherwise
//         2. nil if success
//            non-nil otherwise
func GetIPv4localRoutes() (Routes, error) {
	return VrfGetRoutesByTid(0, nl.FAMILY_V4, RTN_LOCAL)
}

// GetIPv6routes returns a slice of IPv6 netlink.Route
// return: 1. slice of netlink.Route if success. All of them are IPv4
//            undetermined otherwise
//         2. nil if success
//            non-nil otherwise
func GetIPv6routes(name string) (Routes, error) {
	return VrfGetRoutesByTid(0, nl.FAMILY_V6, RTN_UNICAST)
}

// AddRoute adds a route
// in: r Pointer to the route to be added
// return: nil if success
//         non-nil otherwise
func AddRoute(r *Route) error {
	r.Table = 0
	return netlink.RouteAdd(r)
}

// DeleteRoute deletes a route
// in: r Pointer to the route to be added
// return: nil if success
//         non-nil otherwise
func DeleteRoute(r *Route) error {
	r.Table = 0
	return netlink.RouteDel(r)
}

// ReplaceRoute replaces the existing route
// The route is added unless it exists.
// in: r Pointer to the route to be added
// return: nil if success
//         non-nil otherwise
func ReplaceRoute(r *Route) error {
	r.Table = 0
	return netlink.RouteReplace(r)
}
