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
	"golang.org/x/sys/unix"
	"syscall"
	"unsafe"
)

const (
	BRCTL_GET_VERSION = iota
	BRCTL_GET_BRIDGES
	BRCTL_ADD_BRIDGE
	BRCTL_DEL_BRIDGE
)

type Bridge struct {
	Link *netlink.Bridge
}

func getBridgeSock() (int, error) {
	return syscall.Socket(unix.AF_LOCAL, unix.SOCK_STREAM, 0)
}

// bridgeModify adds or deletes a bridge
// in: name Name of the bridge to be added or deleted
//     op Add (true) to add a bridge whose name is `name'
//        Delete (false) to delete a bridge whose name is `name'
//     up Up (true) to bring up the bridge after addition
// return: 1. Pointer to bridge if it is successfully added
//            nil otherwise
//         2. nil if success
//            non-nil otherwise
func bridgeModify(name string, op bool, up bool) (*Bridge, error) {
	banner := fmt.Sprintf("BridgeAdd(%s): ", name)
	var arg [3]uint64
	var brName [unix.IFNAMSIZ]byte

	s, err := getBridgeSock()
	if err != nil {
		fmt.Errorf("%ssocket(): %v\n", banner, err)
	}
	defer syscall.Close(s)

	copy(brName[:unix.IFNAMSIZ-1], name)
	if op == Add {
		arg[0] = BRCTL_ADD_BRIDGE
	} else {
		arg[0] = BRCTL_DEL_BRIDGE
	}
	arg[1] = uint64(uintptr(unsafe.Pointer(&brName)))
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(s),
		unix.SIOCSIFBR, uintptr(unsafe.Pointer(&arg)))
	if errno == 0 {
		if op == Add {
			if br, err := BridgeGetByName(name); err == nil {
				if up {
					return br, br.IfUp()
				}
			} else {
				return nil, err
			}
		} else {
			return nil, nil
		}
	}
	return nil, fmt.Errorf("%s%v\n", banner, errno)
}

// BridgeAdd adds a bridge whose name is `name'
// in: name Name of the bridge to be added
//     up Up (true) to bring up the bridge after addition
// return: 1. Pointer to bridge if it is successfully added
//            nil otherwise
//         2. nil if success
//            non-nil otherwise
func BridgeAdd(name string, up bool) (*Bridge, error) {
	return bridgeModify(name, Add, up)
}

// BridgeDelete deletes a bridge whose name is `name'
// in: name Name of the bridge to be deleted
// return: nil if success
//         non-nil otherwise
func BridgeDelete(name string) error {
	banner := fmt.Sprintf("BridgeDelete(%s): ", name)
	br, err := BridgeGetByName(name)
	if err != nil {
		return fmt.Errorf("%sBridgeGetByName(): %v", banner, err)
	}
	err = br.IfDown()
	if err != nil {
		return fmt.Errorf("%sLinkSetDown(): %v", banner, err)
	}
	_, err = bridgeModify(name, Del, Down)
	return err
}

// BridgeGetByName returns a pointer to Bridge if bridge
// whose name is `name' exists.
// in: name Name of the bridge to be examined
// return: 1. Pointer to bridge if bridge whose name is `name' exists
//            nil otherwise
//         2. nil if bridge whose name is `name' exists
//            non-nil otherwise
func BridgeGetByName(name string) (*Bridge, error) {
	if l, err := netlink.LinkByName(name); err == nil {
		switch l := l.(type) {
		case *netlink.Bridge:
			return &Bridge{Link: l}, nil
		default:
			return nil, fmt.Errorf("BridgeGetByName(%s): not a bridge", name)
		}
	} else {
		return nil, fmt.Errorf("BridgeGetByName(%s): %v", name, err)
	}
}

// BridgeGetByName returns a pointer to Bridge if bridge
// whose ifindex is `i' exists
// in: i Ifindex of the bridge to be examined
// return: 1. Pointer to bridge if bridge whose ifindex is `i' exists
//            nil otherwise
//         2. nil if bridge whose ifindex is `i' exists
//            non-nil otherwise
func BridgeGetByIndex(i int) (*Bridge, error) {
	if l, err := netlink.LinkByIndex(i); err == nil {
		switch l := l.(type) {
		case *netlink.Bridge:
			return &Bridge{Link: l}, nil
		default:
			return nil, fmt.Errorf("BridgeGetByIndex(%d): not a bridge", i)
		}
	} else {
		return nil, fmt.Errorf("BridgeGetByIndex(%d): %v", i, err)
	}
}

// BridgeList returns a slice of Bridge
// return: 1. Slice of Bridge if success
//         2. nil if bridge whose ifindex is `i' exists
//            non-nil otherwise
func BridgeList() ([]Bridge, error) {
	banner := "BridgeList(): "
	var brs []Bridge

	ll, err := netlink.LinkList()
	if err != nil {
		return nil, fmt.Errorf("%sLinkList(): %v", banner, err)
	}
	for _, l := range ll {
		if l.Type() == "bridge" {
			brs = append(brs, Bridge{Link: l.(*netlink.Bridge)})
		}
	}
	return brs, nil
}

// BridgeIfExists returns true if bridge `name' exists
// return: 1. true if bridge `name' exists
//            false otherwise
//         2. nil if bridge whose name is `name' exists
//            non-nil otherwise
func BridgeIfExists(name string) (bool, error) {
	return ifExists(name, &netlink.Bridge{})
}

// BridgeBindIf add interface `ifName' to bridge `brName'
// return: nil if bridge whose name is `name' exists
//         non-nil otherwise
func BridgeBindIf(brName, ifName string) error {
	banner := fmt.Sprintf("BridgeBindIf(%s, %s): ", brName, ifName)
	if br, err := BridgeGetByName(brName); err == nil {
		return br.BindIf(ifName)
	} else {
		return fmt.Errorf("%sBridgeGetByName(): %v", banner, err)
	}
	return nil
}

// Name returns the name of this bridge
func (br *Bridge) Name() string {
	return br.Link.Attrs().Name
}

// Ifup brings up this bridge interface
func (br *Bridge) IfUp() error {
	return netlink.LinkSetUp(br.Link)
}

// Ifup brings down this bridge interface
// return: nil if the bridge whose name is `name' exists
//         non-nil otherwise
func (br *Bridge) IfDown() error {
	return netlink.LinkSetDown(br.Link)
}

// BindIf adds interface `ifName' to this bridge
// return: nil if success
//         non-nil otherwise
func (br *Bridge) BindIf(ifName string) error {
	banner := fmt.Sprintf("BindIf(%s, %s): ", br.Name(), ifName)
	if l, err := netlink.LinkByName(ifName); err == nil {
		return netlink.LinkSetMaster(l, br.Link)
	} else {
		return fmt.Errorf("%sLinkByName(): %v", banner, err)
	}
	return nil
}
