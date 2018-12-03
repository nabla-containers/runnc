// Copyright (c) 2018, IBM
// Author(s): Brandon Lum, Ricardo Koller, Dan Williams
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

package runnc_cont

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/nabla-containers/runnc/nabla-lib/network"
)

type rumpArgsNetwork struct {
	If     string `json:"if"`
	Cloner string `json:"cloner"`
	Type   string `json:"type"`
	Method string `json:"method"`
	Addr   string `json:"addr"`
	Mask   string `json:"mask"`
	Gw     string `json:"gw"`
}

type rumpArgsBlock struct {
	Source string `json:"source"`
	Path   string `json:"path"`
	Fstype string `json:"fstype"`
	Mount  string `json:"mountpoint"`
}

type rumpArgs struct {
	Cmdline string          `json:"cmdline"`
	Net     rumpArgsNetwork `json:"net"`
	Blk     *rumpArgsBlock  `json:"blk,omitempty"`
	Env     []string        `json:"env,omitempty"`
	Cwd     string          `json:"cwd,omitempty"`
	Mem     string          `json:"mem,omitempty"`
}

// Overwrite the rumprum args marshalling since rump expects multiple env
// variables to be passed in a weird way.
func (ra *rumpArgs) MarshalJSON() ([]byte, error) {
	// Create duplicate env variables due to consumption method of rump that
	// requires duplicate json keys.
	env := ra.Env
	type EnvAlias struct {
		Env string `json:"env,omitempty"`
	}

	addString := ""
	type CAlias struct {
		C string `json:"c,omitempty"`
	}

	for _, v := range env {
		vb, err := json.Marshal(&EnvAlias{v})
		if err != nil {
			return nil, err
		}
		addString += string(vb[1:len(vb)-1]) + ","
	}

	// Marshal rest of the struct minus Env
	type Alias rumpArgs
	alias := &struct {
		*Alias
	}{
		Alias: (*Alias)(ra),
	}

	alias.Env = nil
	otherBytes, err := json.Marshal(alias)
	if err != nil {
		return nil, err
	}
	alias.Env = env

	// Put bytes together
	modified := make([]byte, 0, len(otherBytes)+len(addString))
	modified = append(modified, otherBytes[:1]...)
	modified = append(modified, []byte(addString)...)
	modified = append(modified, otherBytes[1:]...)

	return modified, nil
}

// CreateRumprunArgs returns the cmdline string for rumprun (a json)
func CreateRumprunArgs(ip net.IP, mask net.IPMask, gw net.IP,
	mountPoint string, envVars []string, cwd string,
	unikernel string, cmdargs []string) (string, error) {

	// XXX: Due to bug in: https://github.com/nabla-containers/runnc/issues/40
	// If we detect a /32 mask, we set it to 1 as a "fix", and hope we are in
	// the same subnet... (working on a fix for mask:0)
	cidr := strconv.Itoa(network.MaskCIDR(mask))
	if cidr == "32" {
		fmt.Printf("WARNING: Changing CIDR from 32 to 1 due to Issue https://github.com/nabla-containers/runnc/issues/40\n")
		cidr = "1"
	}

	net := rumpArgsNetwork{
		If:     "ukvmif0",
		Cloner: "True",
		Type:   "inet",
		Method: "static",
		Addr:   ip.String(),
		Mask:   cidr,
		Gw:     gw.String(),
	}

	cmdline := append([]string{unikernel}, cmdargs...)
	ra := &rumpArgs{
		Cwd:     cwd,
		Cmdline: strings.Join(cmdline, " "),
		Net:     net,
	}
	if mountPoint != "" {
		block := rumpArgsBlock{
			Source: "etfs",
			Path:   "/dev/ld0a",
			Fstype: "blk",
			Mount:  mountPoint,
		}
		ra.Blk = &block
	}

	if len(envVars) > 0 {
		ra.Env = envVars
	}

	b, err := json.Marshal(ra)
	if err != nil {
		return "", fmt.Errorf("error with rumprun json: %v", err)
	}

	return string(b), nil
}
