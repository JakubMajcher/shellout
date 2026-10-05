// SPDX-License-Identifier: GPL-3.0-or-later

package sysinfo

import (
	"strings"
	"testing"
)

func TestParseOSRelease(t *testing.T) {
	cases := map[string]struct{ in, want string }{
		"arch": {`NAME="Arch Linux"
PRETTY_NAME="Arch Linux"
ID=arch
`, "Arch Linux"},
		"ubuntu": {`PRETTY_NAME="Ubuntu 24.04.1 LTS"
NAME="Ubuntu"
VERSION_ID="24.04"
`, "Ubuntu 24.04.1 LTS"},
		"alpine": {`NAME="Alpine Linux"
ID=alpine
VERSION_ID=3.20.3
PRETTY_NAME="Alpine Linux v3.20"
`, "Alpine Linux v3.20"},
		"no pretty name": {`NAME='Some Linux'
ID=some
`, "Some Linux"},
		"empty": {"", "unknown"},
	}
	for name, c := range cases {
		if got := ParseOSRelease(strings.NewReader(c.in)); got != c.want {
			t.Errorf("%s: got %q, want %q", name, got, c.want)
		}
	}
}

func TestFlavor(t *testing.T) {
	if got := Flavor(true, "sed (GNU sed) 4.9\n", "/usr/bin/sed"); got != "GNU" {
		t.Errorf("GNU: got %s", got)
	}
	if got := Flavor(false, "", "/bin/busybox"); got != "BusyBox" {
		t.Errorf("BusyBox: got %s", got)
	}
	if got := Flavor(false, "sed: illegal option -- -", "/usr/bin/sed"); got != "BSD" {
		t.Errorf("BSD: got %s", got)
	}
}

func TestShellPath(t *testing.T) {
	if ShellPath("") != "/bin/sh" || ShellPath("/bin/zsh") != "/bin/zsh" {
		t.Fatal("ShellPath fallback broken")
	}
}

func TestCollectFillsEveryField(t *testing.T) {
	info := Collect()
	for name, v := range map[string]string{
		"OS": info.OS, "Arch": info.Arch, "Distro": info.Distro,
		"Shell": info.Shell, "ShellPath": info.ShellPath, "Flavor": info.Flavor,
	} {
		if v == "" {
			t.Errorf("%s is empty", name)
		}
	}
}
