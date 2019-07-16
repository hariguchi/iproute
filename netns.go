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
	netns "github.com/hariguchi/go_netns"
	"github.com/vishvananda/netlink"
)

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
