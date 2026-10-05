// SPDX-License-Identifier: GPL-3.0-or-later

package main

import "testing"

func TestParseArgs(t *testing.T) {
	cases := []struct {
		args []string
		want invocation
	}{
		{[]string{"list", "open", "ports"}, invocation{request: "list open ports"}},
		{[]string{"-p", "local", "list", "ports"}, invocation{profile: "local", request: "list ports"}},
		{[]string{"find", "files", "-size", "+1G"}, invocation{request: "find files -size +1G"}},
		{[]string{"init", "zsh"}, invocation{isInit: true, initShell: "zsh"}},
		{[]string{"init", "tcsh"}, invocation{isInit: true, initShell: "tcsh"}},
		{[]string{"init", "a", "git", "repo", "here"}, invocation{request: "init a git repo here"}},
		{[]string{"init"}, invocation{request: "init"}},
		{[]string{"--version"}, invocation{version: true}},
		{[]string{"-h"}, invocation{help: true}},
		{[]string{"--help"}, invocation{help: true}},
		{nil, invocation{}},
	}
	for _, c := range cases {
		got, err := parseArgs(c.args)
		if err != nil {
			t.Errorf("%v: unexpected error %v", c.args, err)
			continue
		}
		if got != c.want {
			t.Errorf("%v: got %+v, want %+v", c.args, got, c.want)
		}
	}
	if _, err := parseArgs([]string{"-x"}); err == nil {
		t.Error("unknown flag must be an error")
	}
	if _, err := parseArgs([]string{"-p"}); err == nil {
		t.Error("-p without a value must be an error")
	}
}
