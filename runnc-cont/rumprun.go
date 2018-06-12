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

package main

import (
	"encoding/json"
	"fmt"
	"github.ibm.com/nabla-containers/nabla-lib/network"
	"net"
	"os"
	"strconv"
	"strings"
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

// CreateRumprunArgs returns the cmdline string for rumprun (a json)
func CreateRumprunArgs(ip net.IP, mask net.IPMask, gw net.IP,
	mountPoint string, envVars []string, cwd string,
	unikernel string, cmdargs string) (string, error) {

	net := rumpArgsNetwork{
		If:     "ukvmif0",
		Cloner: "True",
		Type:   "inet",
		Method: "static",
		Addr:   ip.String(),
		Mask:   strconv.Itoa(network.MaskCIDR(mask)),
		Gw:     gw.String(),
	}

	ra := make(map[string]interface{})
	cmdline := []string{unikernel, cmdargs}
	ra["cwd"] = cwd
	ra["cmdline"] = strings.Join(cmdline, " ")
	ra["net"] = net
	if mountPoint != "" {
		block := rumpArgsBlock{
			Source: "etfs",
			Path:   "/dev/ld0a",
			Fstype: "blk",
			Mount:  mountPoint,
		}
		ra["blk"] = block
	}

	// XXX: Unfortunately the rumprun JSON takes multiple "env" keys which
	// is not valid JSON (at least here in golang). So for now, just take
	// the first one (if any).
	if len(envVars) > 0 {
		ra["env"] = envVars[0]
	}

	if len(envVars) > 1 {
		fmt.Fprintf(os.Stderr,
			"All -env values after the first will be ignored.\n")
	}

	b, err := json.Marshal(ra)
	if err != nil {
		return "", fmt.Errorf("error with rumprun json: %v", err)
	}

	return string(b), nil
}
