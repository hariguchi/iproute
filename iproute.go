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
	"bytes"
	"fmt"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"net"
	"regexp"
)

const (
	DefaultTxQlen      int = 1000
	DefaultTxQueues    int = 1
	DefaultRxQueues    int = 1
	DefaultMTU         int = 1500
	FAMILY_ALL             = netlink.FAMILY_ALL
	FAMILY_V4              = netlink.FAMILY_V4
	FAMILY_V6              = netlink.FAMILY_V6
	FAMILY_MPLS            = netlink.FAMILY_MPLS
	FLAG_ONLINK            = netlink.FLAG_ONLINK
	FLAG_PERVASIVE         = netlink.FLAG_PERVASIVE
	RT_FILTER_PROTOCOL     = netlink.RT_FILTER_PROTOCOL
	RT_FILTER_SCOPE        = netlink.RT_FILTER_SCOPE
	RT_FILTER_TYPE         = netlink.RT_FILTER_TYPE
	RT_FILTER_TOS          = netlink.RT_FILTER_TOS
	RT_FILTER_IIF          = netlink.RT_FILTER_IIF
	RT_FILTER_OIF          = netlink.RT_FILTER_OIF
	RT_FILTER_DST          = netlink.RT_FILTER_DST
	RT_FILTER_SRC          = netlink.RT_FILTER_SRC
	RT_FILTER_GW           = netlink.RT_FILTER_GW
	RT_FILTER_TABLE        = netlink.RT_FILTER_TABLE
	RTN_UNSPEC             = unix.RTN_UNSPEC
	RTN_UNICAST            = unix.RTN_UNICAST
	RTN_LOCAL              = unix.RTN_LOCAL
	RTN_BROADCAST          = unix.RTN_BROADCAST
	RTN_ANYCAST            = unix.RTN_ANYCAST
	RTN_MULTICAST          = unix.RTN_MULTICAST
	RTN_BLACKHOLE          = unix.RTN_BLACKHOLE
	RTN_UNREACHABLE        = unix.RTN_UNREACHABLE
	RTN_PROHIBIT           = unix.RTN_PROHIBIT
	RTN_THROW              = unix.RTN_THROW
	RTN_NAT                = unix.RTN_NAT
	RTN_XRESOLVE           = unix.RTN_XRESOLVE
	SCOPE_UNIVERSE         = netlink.SCOPE_UNIVERSE
	SCOPE_SITE             = netlink.SCOPE_SITE
	SCOPE_LINK             = netlink.SCOPE_LINK
	SCOPE_HOST             = netlink.SCOPE_HOST
	SCOPE_NOWHERE          = netlink.SCOPE_NOWHERE
)

const (
	Add  = true
	Del  = false
	Up   = true
	Down = false
	Self = true
	Peer = false
)

type Link = netlink.Link
type NHinfo = netlink.NexthopInfo
type Route = netlink.Route
type Routes = []netlink.Route
type IPs []net.IP

// Len returns the number of elements in `a'
// It is used in sort.Sort(IPs)
func (a IPs) Len() int {
	return len(a)
}

// Swap swaps the contents of a[i] and a[j]
// It is used in sort.Sort(IPs)
func (a IPs) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// Less returns true if a[i] < a[j]; it returns false otherwise.
func (a IPs) Less(i, j int) bool {
	l := len(a[i])
	if l != len(a[j]) {
		panic(fmt.Sprintf("IPs.Less(%d, %d): %d, %d", i, j, l, len(a[j])))
	}
	if bytes.Compare(a[i], a[j]) < 0 {
		return true
	}
	return false
}

// IPNetEqual returns true if the contents of two net.IPNet instances are
// identical
// in: a Pointer to net.IPNet
//     b Pointer to net.IPNet
// return: nil if success
//         non-nil otherwise
func IPNetEqual(a, b *net.IPNet) bool {
	sa, _ := a.Mask.Size()
	sb, _ := b.Mask.Size()
	if a.IP.Equal(b.IP) && sa == sb {
		return true
	} else {
		return false
	}
}

func isThisInError(s string, err error) bool {
	re := regexp.MustCompile(s)
	if re.MatchString(fmt.Sprint(err)) {
		return true
	}
	return false
}

func IsExist(err error) bool {
	return isThisInError(`(already|file) exists`, err)
}

func IsNotFound(err error) bool {
	return isThisInError(`not found`, err)
}
