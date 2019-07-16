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
	"regexp"
	"syscall"
	"unsafe"
)

const (
	BRCTL_GET_VERSION = iota
	BRCTL_GET_BRIDGES
	BRCTL_ADD_BRIDGE
	BRCTL_DEL_BRIDGE
)

const (
	add = true
	del = false
)

type Bridge struct {
	Link *netlink.Bridge
}

func getBridgeSock() (int, error) {
	return syscall.Socket(unix.AF_LOCAL, unix.SOCK_STREAM, 0)
}

func bridgeModify(name string, op bool) error {
	banner := fmt.Sprintf("BridgeAdd(%s): ", name)
	var arg [3]uint64
	var brName [unix.IFNAMSIZ]byte

	s, err := getBridgeSock()
	if err != nil {
		fmt.Errorf("%ssocket(): %v\n", banner, err)
	}
	defer syscall.Close(s)

	copy(brName[:unix.IFNAMSIZ-1], name)
	if op == add {
		arg[0] = BRCTL_ADD_BRIDGE
	} else {
		arg[0] = BRCTL_DEL_BRIDGE
	}
	arg[1] = uint64(uintptr(unsafe.Pointer(&brName)))
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(s),
		unix.SIOCSIFBR, uintptr(unsafe.Pointer(&arg)))
	if errno == 0 {
		return nil
	}
	return fmt.Errorf("%s%v\n", banner, errno)
}

func BridgeAdd(name string) error {
	return bridgeModify(name, add)
}

func BridgeDelete(name string) error {
	banner := fmt.Sprintf("BridgeDelete(%s): ", name)
	br, err := BridgeGetByName(name)
	if err != nil {
		return fmt.Errorf("%sBridgeGetByName(): %v", banner, err)
	}
	err = netlink.LinkSetDown(br.Link)
	if err != nil {
		return fmt.Errorf("%sLinkSetDown(): %v", banner, err)
	}
	return bridgeModify(name, del)
}

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

func BridgeAddIfByName(brName, ifName string) error {
	banner := fmt.Sprintf("BridgeAddIfByName(%s, %s): ", brName, ifName)
	if br, err := BridgeGetByName(brName); err == nil {
		return br.AddIfByName(ifName)
	} else {
		return fmt.Errorf("%sBridgeGetByName(): %v", banner, err)
	}
	return nil
}

func BridgeDeleteIfByName(ifName string) error {
	banner := fmt.Sprintf("BridgeDeleteIfByName(%s): ", ifName)
	if l, err := netlink.LinkByName(ifName); err == nil {
		return netlink.LinkSetNoMaster(l)
	} else {
		return fmt.Errorf("%sLinkByName(): %v", banner, err)
	}
	return nil
}

func (br *Bridge) Name() string {
	return br.Link.Attrs().Name
}

func (br *Bridge) IfUp() error {
	return netlink.LinkSetUp(br.Link)
}

func (br *Bridge) IfDown() error {
	return netlink.LinkSetDown(br.Link)
}

func (br *Bridge) AddIfByName(ifName string) error {
	banner := fmt.Sprintf("AddIfByName(%s, %s): ", br.Name(), ifName)
	if l, err := netlink.LinkByName(ifName); err == nil {
		return netlink.LinkSetMaster(l, br.Link)
	} else {
		return fmt.Errorf("%sLinkByName(): %v", banner, err)
	}
	return nil
}

func (br *Bridge) IsNotExist(err error) bool {
	re := regexp.MustCompile(`not found`)
	if re.MatchString(fmt.Sprint(err)) {
		return true
	}
	return false
}
