// Copyright (c) 2018, IBM
// Author(s): Brandon Lum, Ricardo Koller
//
// SPDX-License-Identifier: ISC
//
// Copyright (c) 2016 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//
// Permission to use, copy, modify, and/or distribute this software for
// any purpose with or without fee is hereby granted, provided that the
// above copyright notice and this permission notice appear in all
// copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL
// WARRANTIES WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE
// AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL
// DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS OF USE, DATA
// OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER
// TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

// +build linux

package network

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// CreateBridge creates and returns a netklink.Bridge
func CreateBridge(bridgeName string) (*netlink.Bridge, error) {
	la := netlink.NewLinkAttrs()
	la.Name = bridgeName
	mybridge := &netlink.Bridge{LinkAttrs: la}
	err := netlink.LinkAdd(mybridge)
	if err != nil {
		return nil, fmt.Errorf("could not add %s: %v", la.Name, err)
	}
	return mybridge, err
}

// MaskCIDR returns a mask CIDR (the 16 in 1.1.1.1/16) for a net.IPMask
func MaskCIDR(mask net.IPMask) int {
	m, _ := mask.Size()
	return int(m)
}

// CreateTapInterface creates a new TAP interface and assignes it ip/mask as
// the new address.
func CreateTapInterface(tapName string, ip net.IP, mask net.IPMask) error {

	err := SetupTunDev()
	if err != nil {
		return err
	}

	// ip tuntap add %s mode tap
	tap := &netlink.Tuntap{
		LinkAttrs: netlink.LinkAttrs{Name: tapName},
		Mode:      netlink.TUNTAP_MODE_TAP}
	err = netlink.LinkAdd(tap)
	if err != nil {
		return err
	}

	// ip addr add %s/%s dev %s
	netstr := fmt.Sprintf("%s/%d", ip.String(), MaskCIDR(mask))
	addr, err := netlink.ParseAddr(netstr)
	if err != nil {
		return err
	}

	netlink.AddrAdd(tap, addr)

	// ip link set dev %s up'
	err = netlink.LinkSetUp(tap)
	if err != nil {
		return err
	}
	return nil
}

// CreateMacvtapInterfaceDocker creates a Macvtap interface associated with
// master (usually "eth0").  Returns the assigned IP/mask and gateway IP
// (previously owned by master) and the MAC of the Macvtap interface that has
// to be used by the unikernel's NIC.
//
// Got the idea of using macvtap's and the fix for the inability to get the
// right index in a network namespace from the Kata containers repository:
// https://github.com/kata-containers/runtime/blob/593bd44f207aa7b21e561184ca1b3fb79da47eb6/virtcontainers/network.go
//
func CreateMacvtapInterfaceDocker(tapName *string, master string) (
	net.IP, net.IP, net.IPMask, string, error) {

	netHandle, err := netlink.NewHandle()
	if err != nil {
		return nil, nil, nil, "", err
	}

	err = SetupTunDev()
	if err != nil {
		return nil, nil, nil, "", err
	}

	masterLink, err := netlink.LinkByName(master)
	if err != nil {
		return nil, nil, nil, "",
			fmt.Errorf("no master interface: %v", err)
	}
	masterLinkAttrs := masterLink.Attrs()

	rand.Seed(time.Now().Unix())

	qlen := masterLinkAttrs.TxQLen
	if qlen <= 0 {
		qlen = 1000
	}

	index := 8192 + rand.Intn(1024)
	name := fmt.Sprintf("macvtap%d", index)

	macvtapLink := &netlink.Macvtap{
		Macvlan: netlink.Macvlan{
			Mode: netlink.MACVLAN_MODE_BRIDGE,
			LinkAttrs: netlink.LinkAttrs{
				Index:       index,
				Name:        name,
				TxQLen:      qlen,
				ParentIndex: masterLinkAttrs.Index,
			},
		},
	}

	err = netHandle.LinkAdd(macvtapLink)
	if err != nil {
		return nil, nil, nil, "",
			fmt.Errorf("Couldn't add newlink: %v", err)
	}

	*tapName = fmt.Sprintf("/dev/tap%d", macvtapLink.Attrs().Index)

	addrs, err := netlink.AddrList(masterLink, netlink.FAMILY_V4)
	if err != nil {
		return nil, nil, nil, "", err
	}
	if len(addrs) == 0 {
		return nil, nil, nil, "", fmt.Errorf("master should have an IP")
	}
	masterAddr := addrs[0]
	masterIP := addrs[0].IPNet.IP
	masterMask := addrs[0].IPNet.Mask

	routes, err := netlink.RouteList(masterLink, netlink.FAMILY_V4)
	if err != nil {
		return nil, nil, nil, "", err
	}
	if len(routes) == 0 {
		return nil, nil, nil, "",
			fmt.Errorf("master should have at least one route")
	}
	// XXX: is the "gateway" always the first route?
	gwAddr := routes[0].Gw

	// ip addr del $INET_STR dev master
	err = netlink.AddrDel(masterLink, &masterAddr)
	if err != nil {
		return nil, nil, nil, "", err
	}

	err = netlink.LinkSetUp(macvtapLink)
	if err != nil {
		return nil, nil, nil, "", err
	}

	err = netlink.LinkSetUp(masterLink)
	if err != nil {
		return nil, nil, nil, "", err
	}

	// The HardwareAddr Attr doesn't automatically get updated
	_macvtapLink, err := netlink.LinkByName(name)
	if err != nil {
		return nil, nil, nil, "", err
	}
	tapMac := _macvtapLink.Attrs().HardwareAddr.String()

	d := fmt.Sprintf("/sys/devices/virtual/net/%s/tap%d/dev",
		name, macvtapLink.Attrs().Index)
	b, err := ioutil.ReadFile(d)
	if err != nil {
		return nil, nil, nil, "", err
	}

	mm := strings.Split(string(b), ":")
	major, err := strconv.Atoi(strings.TrimSpace(mm[0]))
	if err != nil {
		return nil, nil, nil, "", err
	}

	minor, err := strconv.Atoi(strings.TrimSpace(mm[1]))
	if err != nil {
		return nil, nil, nil, "", err
	}

	fmt.Printf("%s = %d %d\n", string(b), major, minor)

	err = unix.Mknod(*tapName, unix.S_IFCHR|0600,
		int(unix.Mkdev(uint32(major), uint32(minor))))
	if err != nil {
		return nil, nil, nil, "", err
	}

	return masterIP, gwAddr, masterMask, tapMac, nil
}

// CreateTapInterfaceDocker creates a new TAP interface and a bridge, adds both
// the TAP and the master link (usually eth0) to the bridge, and unsets the IP
// of the master link to be used by the unikernel NIC.  Returns the assigned
// IP/mask and gateway IP.
func CreateTapInterfaceDocker(tapName string, master string) (
	net.IP, net.IP, net.IPMask, error) {

	masterLink, err := netlink.LinkByName(master)
	if err != nil {
		return nil, nil, nil,
			fmt.Errorf("no master interface: %v", err)
	}

	addrs, err := netlink.AddrList(masterLink, netlink.FAMILY_V4)
	if err != nil {
		return nil, nil, nil, err
	}
	if len(addrs) == 0 {
		return nil, nil, nil, fmt.Errorf("master should have an IP")
	}
	masterAddr := addrs[0]
	masterIP := addrs[0].IPNet.IP
	masterMask := addrs[0].IPNet.Mask

	routes, err := netlink.RouteList(masterLink, netlink.FAMILY_V4)
	if err != nil {
		return nil, nil, nil, err
	}
	if len(routes) == 0 {
		return nil, nil, nil,
			fmt.Errorf("master should have at least one route")
	}
	// XXX: is the "gateway" always the first route?
	gwAddr := routes[0].Gw

	err = SetupTunDev()
	if err != nil {
		return nil, nil, nil, err
	}

	// ip tuntap add tap100 mode tap
	tap := &netlink.Tuntap{
		LinkAttrs: netlink.LinkAttrs{Name: tapName},
		Mode:      netlink.TUNTAP_MODE_TAP}
	err = netlink.LinkAdd(tap)
	if err != nil {
		return nil, nil, nil, err
	}

	// ip link set dev tap100 up'
	err = netlink.LinkSetUp(tap)
	if err != nil {
		return nil, nil, nil, err
	}

	// ip addr del $INET_STR dev master
	err = netlink.AddrDel(masterLink, &masterAddr)
	if err != nil {
		return nil, nil, nil, err
	}

	br0, err := CreateBridge("br0")
	if err != nil {
		return nil, nil, nil, err
	}

	netlink.LinkSetMaster(masterLink, br0)
	netlink.LinkSetMaster(tap, br0)

	// ip link set dev br0 up'
	err = netlink.LinkSetUp(br0)
	if err != nil {
		return nil, nil, nil, err
	}

	return masterIP, gwAddr, masterMask, nil
}

// SetupTunDev sets up the /dev/net/tun device if it doesn't exists
func SetupTunDev() error {
	// Check if tun device exists and create it if required
	if err := verifyTunDevice(); err != nil {
		if err = createTunDevice(); err != nil {
			return fmt.Errorf("Unable to create /dev/net/tun: %v",
				err)
		}
	} else {
		return nil
	}

	// Make sure that it is the correct device we're talking to
	if err := verifyTunDevice(); err != nil {
		return err
	}

	return nil
}

// createTunDevice Create directory /dev/net and create tun char device M 10 m
// 200
func createTunDevice() error {
	// Check for directory /dev/net
	devNetInfo, err := os.Stat("/dev/net")
	if err == nil {
		// Check that it is a directory
		if devNetInfo.Mode()&os.ModeDir == 0 {
			return fmt.Errorf("/dev/net is not a directory")
		}
		// Check if dir did not exist, create it
	} else if os.IsNotExist(err) {
		err = os.Mkdir("/dev/net", 0755)
		if err != nil {
			return err
		}
	} else {
		return err
	}

	// Casting to int is safe since it preserves MSB, Mkdev produces 64-bit and
	// it is backward compatible.
	// ref: https://github.com/golang/sys/blob/master/unix/dev_linux.go
	err = unix.Mknod("/dev/net/tun", unix.S_IFCHR|0666, int(unix.Mkdev(10, 200)))
	if err != nil {
		return err
	}

	return nil
}

// verifyTunDevice verifies the /dev/net/tun device that it is char device M 10 m 200
func verifyTunDevice() error {
	var st unix.Stat_t
	err := unix.Stat("/dev/net/tun", &st)
	if err != nil {
		return err
	}

	// File exists, check character device name
	maj := unix.Major(uint64(st.Rdev))
	min := unix.Minor(uint64(st.Rdev))
	if maj != 10 || min != 200 {
		return fmt.Errorf("Expected /dev/net/tun to have M/m %d/%d, got %d/%d", 10, 200, maj, min)
	}

	return nil
}
