package iproute

/*
Route:

Scope:
netlink.Scope
const (
    SCOPE_UNIVERSE Scope = unix.RT_SCOPE_UNIVERSE
    SCOPE_SITE     Scope = unix.RT_SCOPE_SITE
    SCOPE_LINK     Scope = unix.RT_SCOPE_LINK
    SCOPE_HOST     Scope = unix.RT_SCOPE_HOST
    SCOPE_NOWHERE  Scope = unix.RT_SCOPE_NOWHERE
)

Protocol:
golang.org/x/sys/unix
  RTPROT_UNSPEC
  RTPROT_REDIRECT
  RTPROT_KERNEL
  RTPROT_BOOT
  RTPROT_STATIC

Type:
golang.org/x/sys/unix
  RTN_UNSPEC
  RTN_UNICAST
  RTN_LOCAL
  RTN_BROADCAST
  RTN_ANYCAST
  RTN_MULTICAST
  RTN_BLACKHOLE
  RTN_UNREACHABLE
  RTN_PROHIBIT
  RTN_THROW
  RTN_NAT
  RTN_XRESOLVE

Flags:
golang.org/x/sys/unix
  RTNH_F_DEAD                          = 0x1
  RTNH_F_PERVASIVE                     = 0x2
  RTNH_F_ONLINK                        = 0x4
  RTNH_F_OFFLOAD                       = 0x8
  RTNH_F_LINKDOWN                      = 0x10
  RTNH_F_UNRESOLVED                    = 0x20

  RTM_F_NOTIFY                         = 0x100
  RTM_F_CLONED                         = 0x200
  RTM_F_EQUALIZE                       = 0x400
  RTM_F_PREFIX                         = 0x800
  RTM_F_LOOKUP_TABLE                   = 0x1000
  RTM_F_FIB_MATCH                      = 0x2000



type Route struct {
    LinkIndex  int           // output interface index
    ILinkIndex int           // input interface index
    Scope      Scope         // see above
    Dst        *net.IPNet
    Src        net.IP
    Gw         net.IP
    MultiPath  []*NexthopInfo
    Protocol   int
    Priority   int          // metric
    Table      int
    Type       int
    Tos        int
    Flags      int
    MPLSDst    *int
    NewDst     Destination
    Encap      Encap
    MTU        int
    AdvMSS     int
    Hoplimit   int
}

type NexthopInfo struct {
    LinkIndex int
    Hops      int               // weight
    Gw        net.IP
    Flags     int
    NewDst    Destination
    Encap     Encap
}

XxxSubscribe()
XxxSubscribeAt()
XxxSubscribeWith()

*/

import (
	"bytes"
	"fmt"
	netns "github.com/hariguchi/go_netns"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netlink/nl"
	"golang.org/x/sys/unix"
	"net"
	"regexp"
)

const (
	DefaultTxQlen   int = 1000
	DefaultTxQueues int = 1
	DefaultRxQueues int = 1
	DefaultMTU      int = 1500
	FAMILY_ALL          = netlink.FAMILY_ALL
	FAMILY_V4           = netlink.FAMILY_V4
	FAMILY_V6           = netlink.FAMILY_V6
	FAMILY_MPLS         = netlink.FAMILY_MPLS
	RTN_UNSPEC          = unix.RTN_UNSPEC
	RTN_UNICAST         = unix.RTN_UNICAST
	RTN_LOCAL           = unix.RTN_LOCAL
	RTN_BROADCAST       = unix.RTN_BROADCAST
	RTN_ANYCAST         = unix.RTN_ANYCAST
	RTN_MULTICAST       = unix.RTN_MULTICAST
	RTN_BLACKHOLE       = unix.RTN_BLACKHOLE
	RTN_UNREACHABLE     = unix.RTN_UNREACHABLE
	RTN_PROHIBIT        = unix.RTN_PROHIBIT
	RTN_THROW           = unix.RTN_THROW
	RTN_NAT             = unix.RTN_NAT
	RTN_XRESOLVE        = unix.RTN_XRESOLVE
)

type Link = netlink.Link
type NHinfo = netlink.NexthopInfo
type Route = netlink.Route
type Routes = []netlink.Route
type IPs []net.IP

type Vrf struct {
	Name  string
	Index int
	Tid   uint32
}

type Veth struct {
	Name   string
	Peer   string
	TxQlen int
	MTU    int
	NtxQs  int
	NrxQs  int
}

type Vlan struct {
	Name string
	Vid  uint16
}

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

// Equal checks if two Vrfs are identical or not.
// in: other: Pointer to `Vrf'
// return: true if vrf `other' has the same name, ifindex, and table id
//         false otherwise.
func (vrf *Vrf) Equal(other *Vrf) bool {
	if vrf.Name == other.Name &&
		vrf.Index == other.Index && vrf.Tid == other.Tid {
		return true
	}
	return false
}

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

// VrfGetLinkByIndex returns a pointer to netlink.Vrf
// whose ifindex is `idx'
// in: idx Ifindex of the target VRF
// return: A pointer to netlink.Vrf whose ifindex is `idx'.
// An error is returned if there is no vrf whose ifindex is `idx'
func VrfGetLinkByIndex(idx int) (*netlink.Vrf, error) {
	if l, err := netlink.LinkByIndex(idx); err == nil {
		switch l := l.(type) {
		case *netlink.Vrf:
			return l, nil
		default:
			return nil, fmt.Errorf("Error: VrfGetByIndex(%d): not VRF", idx)
		}
	} else {
		return nil, err
	}
}

// VrfGetLinkByName returns a pointer to netlink.Vrf whose name is `name'
// in: name Name of VRF
// return: 1. Pointer to netlink.Vrf associated with `name'
//            undetermined otherwise.
//         2. nil if there is a VRF whose name is `name'
//            non-nil otherwise
func VrfGetLinkByName(name string) (*netlink.Vrf, error) {
	if l, err := netlink.LinkByName(name); err == nil {
		switch l := l.(type) {
		case *netlink.Vrf:
			return l, nil
		default:
			return nil, fmt.Errorf("Error: VrfGetByName(%s): not VRF", name)
		}
	} else {
		return nil, err
	}
}

// VrfGetByIndex returns a pointer to Vrf whose ifindex is `idx'
// in: idx Ifindex of the target VRF
// return: 1. Pointer to Vrf if the VRF whose ifindex is `idx' exists
//            nil otherwise
//         2. nil if the VRF whose ifindex is `idx' exists
//            non-nil otherwise
func VrfGetByIndex(idx int) (*Vrf, error) {
	if l, err := VrfGetLinkByIndex(idx); err == nil {
		return &Vrf{
			Name:  l.Attrs().Name,
			Index: l.Attrs().Index,
			Tid:   l.Table,
		}, nil
	} else {
		return nil, err
	}
}

// VrfGetByName returns a pointer to Vrf whose name is `name'
// in: name Name of the target VRF
// return 1. Pointer to Vrf if the VRF whose name is `name' exists
//           nil otherwise
//        2. nil if  the VRF whose name is `name' exists
//           non-nil otherwise
func VrfGetByName(name string) (*Vrf, error) {
	if l, err := VrfGetLinkByName(name); err == nil {
		return &Vrf{
			Name:  l.Attrs().Name,
			Index: l.Attrs().Index,
			Tid:   l.Table,
		}, nil
	} else {
		return nil, err
	}
}

// VrfAdd adds the VRF whose name is `name' and whose table id is `tid'
// in: name Name of VRF
//     tid Table ID for VRF `name'
// return: nil if success
//         non-nil otherwise
func VrfAdd(name string, tid uint32) error {
	return netlink.LinkAdd(&netlink.Vrf{
		LinkAttrs: netlink.LinkAttrs{Name: name},
		Table:     uint32(tid),
	})
}

// VrfDelete deletes the VRF whose name is `name'
// in: name Name of VRF
// return: nil if success
//         non-nil otherwise
func VrfDelete(name string) error {
	return LinkDel(name)
}

// VrfBindIntf binds an interface to a VRF
// in: vrfName Name of VRF
//     ifName Name of interface to be bound to VRF `vrfName'
// return: nil if success
//         non-nil otherwise
func VrfBindIntf(vrfName, ifName string) error {
	if vrf, err := VrfGetLinkByName(vrfName); err == nil {
		if l, err := netlink.LinkByName(ifName); err == nil {
			return netlink.LinkSetMasterByIndex(l, vrf.Attrs().Index)
		} else {
			return err
		}
	} else {
		return err
	}
}

// VrfBindIntf binds an interface to a VRF
// in: vrfName Name of VRF
//     ifName Name of interface to be bound to VRF `vrfName'
// return: nil if success
//         non-nil otherwise
func VrfUnbindIntf(ifName string) error {
	if l, err := netlink.LinkByName(ifName); err == nil {
		return netlink.LinkSetNoMaster(l)
	} else {
		return err
	}
}

// VrfIndexOf returns ifindex associated with the VRF
// whose name is `name'
// in: name Name of VRF
// return 1. Ifindex (> 0) if VRF whose name is `name' exists
//           -1 otherwise
//        2. nil if success
//           non-nil otherwise
func VrfIndexOf(name string) (int, error) {
	if l, err := netlink.LinkByName(name); err == nil {
		return l.Attrs().MasterIndex, nil
	} else {
		return -1, err
	}
}

// VrfOf returns the pointer to Vrf associated with the VRF
// whose name is `name'
// in: name Name of VRF
// return 1. Poiner to Vrf if VRF whose name is `name' exists
//           nil otherwise
//        2. nil if success
//           non-nil otherwise
func VrfOf(name string) (*Vrf, error) {
	if l, err := netlink.LinkByName(name); err == nil {
		return VrfGetByIndex(l.Attrs().MasterIndex)
	} else {
		return nil, err
	}
}

// VrfGetRoutesByTid returns a slise of netlink.Route whose
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
	filterMask := netlink.RT_FILTER_TABLE | netlink.RT_FILTER_TYPE
	return netlink.RouteListFiltered(family, routeFilter, filterMask)
}

// VrfGetRoutesByName returns a slise of netlink.Route belonging to the VRF
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
		return VrfGetRoutesByTid(int(vrf.Tid), family, tblType)
	} else {
		errMsg := fmt.Sprintf("Error: VrfGetRoutesByName(%s): ", vrf.Name)
		return nil, fmt.Errorf(errMsg+"%v", err)
	}
}

// VrfGetIPv4routesByName returns a slise of IPv4 netlink.Route
// belonging to the VRF whose nanme is `name'
// in: name Name of VRF
// return: 1. slice of netlink.Route if success. All of them are IPv4
//            undetermined otherwise
//         2. nil if success
//            non-nil otherwise
func VrfGetIPv4routesByName(name string) (Routes, error) {
	if vrf, err := VrfGetByName(name); err == nil {
		return VrfGetRoutesByTid(int(vrf.Tid), nl.FAMILY_V4, RTN_UNICAST)
	} else {
		return Routes{}, err
	}
}

// VrfGetIPv4localRoutesByName returns a slise of IPv4 netlink.Route of
// local routes belonging to the VRF whose nanme is `name'
// in: name Name of VRF
// return: 1. slice of netlink.Route if success. All of them are IPv4
//            undetermined otherwise
//         2. nil if success
//            non-nil otherwise
func VrfGetIPv4localRoutes(vrf string) (Routes, error) {
	return VrfGetRoutesByName(vrf, nl.FAMILY_V4, RTN_LOCAL)
}

// VrfGetIPv6routesByName returns a slise of IPv6 netlink.Route
// belonging to the VRF whose nanme is `name'
// in: name Name of VRF
// return: 1. slice of netlink.Route if success. All of them are IPv4
//            undetermined otherwise
//         2. nil if success
//            non-nil otherwise
func VrfGetIPv6routesByName(name string) (Routes, error) {
	if vrf, err := VrfGetByName(name); err == nil {
		return VrfGetRoutesByTid(int(vrf.Tid), nl.FAMILY_V6, RTN_UNICAST)
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
	errMsg := fmt.Sprintf("Error: VrfAddRouteByName(%s, %v): ", name, r)
	if vrf, err := VrfGetByName(name); err == nil {
		r.Table = int(vrf.Tid)
		return netlink.RouteAdd(r)
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
	errMsg := fmt.Sprintf("ERROR: VrfAddRouteByName(%s, %v): ", name, r)
	if vrf, err := VrfGetByName(name); err == nil {
		r.Table = int(vrf.Tid)
		return netlink.RouteReplace(r)
	} else {
		return fmt.Errorf(errMsg+"VrfGetByName(): %v", err)
	}
}

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
		v.Flags |= int(netlink.FLAG_ONLINK)
	case *NHinfo:
		v.Flags |= int(netlink.FLAG_ONLINK)
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
		v.Flags &^= int(netlink.FLAG_ONLINK)
	case *NHinfo:
		v.Flags &^= int(netlink.FLAG_ONLINK)
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
		v.Flags |= int(netlink.FLAG_PERVASIVE)
	case *NHinfo:
		v.Flags |= int(netlink.FLAG_PERVASIVE)
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
		v.Flags &^= int(netlink.FLAG_PERVASIVE)
	case *NHinfo:
		v.Flags &^= int(netlink.FLAG_PERVASIVE)
	default:
		err = fmt.Errorf("ERROR: ClearPervasive(%v): wrong type", v)
	}
	return err
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

// IfSetNS bind an interface to a network namespace
// in: ifName Name of the interface to be bound
//     nsName Name of the network namespace to bind `ifName'
// return: nil if success
//         non-nil otherwise
func IfSetNS(ifName, nsName string) error {
	var (
		err error
		l   netlink.Link
		h   netns.NsHandle
	)
	l, err = netlink.LinkByName(ifName)
	if err != nil {
		return fmt.Errorf("IfSetNS(): LinkByName(%s): %v", ifName, err)
	}
	h, err = netns.GetHandleByName(nsName)
	if err != nil {
		return fmt.Errorf("IfSetNs(): GetHandleByName(%s): %v", nsName, err)
	}
	return netlink.LinkSetNsFd(l, int(h))
}

// IfSetNSbyPid bind an interface to a network namespace
// in: ifName Name of the interface to be bound
//     pid Profess ID of the network namespace to bind `ifName'
// return: nil if success
//         non-nil otherwise
func IfSetNSbyPid(ifName string, pid int) error {
	l, err := netlink.LinkByName(ifName)
	if err != nil {
		return fmt.Errorf("IfSetNSbyPid(): LinkByName(%s): %v", ifName, err)
	}
	return netlink.LinkSetNsPid(l, pid)
}

// IfUnsetNS unbind an interface from a network namespace
// in: ifName Name of the interface to be unbound
//     nsName Name of the network namespace to unbind `ifName'
// return: nil if success
//         non-nil otherwise
func IfUnsetNS(ifName, nsName string) error {
	var (
		errMsg string
		ns     netns.NsDesc
	)
	//
	// save the current namespace handle
	//
	h, err := netns.GetMyHandle()
	if err != nil {
		return fmt.Errorf("IfUnsetNS(%s): GetMyHandle(): %v", nsName, err)
	}
	//
	// switch to namespace `nsName'
	//
	ns, err = netns.SetByName(nsName)
	if err != nil {
		return fmt.Errorf("IfUnsetNS(): namespace %s: %v", nsName, err)
	}
	defer ns.Close()

	//
	// delete interface from this namespace
	//
	err = IfSetNSbyPid(ifName, 1)
	if err != nil {
		errMsg = fmt.Sprintf("IfSetNSbyPid(%s, 1): %v. ", ifName, err)
	}
	//
	// back to original namespace
	//
	err = netns.SetByHandle(h)
	if err != nil {
		errMsg += fmt.Sprintf(
			"IfUnsetNS(%s): failed to switch back namespace: %v", nsName, err)
	}
	if errMsg == "" {
		return nil
	}
	return fmt.Errorf(errMsg)
}

// SetNS bind veth `v' to network namespace `nsName'
// in: nsName Name of the network namespace to bind `v'
// return nil if success
//        non-nil otherwise
func (v *Veth) SetNS(nsName string, up bool) error {
	return IfSetNS(v.Name, nsName)
}

// SetNSbyPid bind veth `v' to network namespace whose process ID is `pid'
// in: nsName Name of the network namespace to bind `v'
// return nil if success
//        non-nil otherwise
func (v *Veth) SetNSbyPid(pid int) error {
	return IfSetNSbyPid(v.Name, pid)
}

// UnsetNS unbinds veth `v' from network namespace whose
// process ID is `pid'
// in: nsName Name of the network namespace to unbind `v'
// return nil if success
//        non-nil otherwise
func (v *Veth) UnsetNS(nsName string) error {
	return IfUnsetNS(v.Name, nsName)
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

// VlanAdd adds a VLAN interface to the master interface
// in: ifName Name of the master interface
//     vlanId VLAN ID
// return: 1. Name of the VLAN interface if success
//            Empty string otherwise
//         2. nil if success
//            non-nil otherwise
func VlanAdd(ifName string, vlanId uint16) (string, error) {
	if l, err := netlink.LinkByName(ifName); err == nil {
		ifName := fmt.Sprintf("%s.%d", ifName, vlanId)
		return ifName, netlink.LinkAdd(&netlink.Vlan{
			netlink.LinkAttrs{
				Name:        ifName,
				ParentIndex: l.Attrs().Index,
			},
			int(vlanId)})
	} else {
		return "", err
	}
}

// VlanDelete deletes the specified VLAN interface
// in: name Name of the VLAN interface to be deleted
// return: nil if success
//         non-nil otherwise
func VlanDelete(name string) error {
	return LinkDel(name)
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
