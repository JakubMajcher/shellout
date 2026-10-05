// SPDX-License-Identifier: GPL-3.0-or-later

// Package sysinfo collects the environment facts sent to the model:
// OS, architecture, distribution, shell and coreutils flavor. Nothing else.
package sysinfo

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Info describes the machine for the system prompt.
type Info struct {
	OS        string // runtime.GOOS
	Arch      string // runtime.GOARCH
	Distro    string // e.g. "Arch Linux", "macOS 15.1"
	Shell     string // base name of ShellPath, e.g. "zsh"
	ShellPath string // $SHELL, or /bin/sh
	Flavor    string // "GNU", "BusyBox" or "BSD"
}

// Collect gathers Info for the current machine.
func Collect() Info {
	sp := ShellPath(os.Getenv("SHELL"))
	return Info{
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Distro:    distro(),
		Shell:     filepath.Base(sp),
		ShellPath: sp,
		Flavor:    detectFlavor(),
	}
}

// ShellPath returns the shell used to run commands.
func ShellPath(shellEnv string) string {
	if shellEnv == "" {
		return "/bin/sh"
	}
	return shellEnv
}

func distro() string {
	if runtime.GOOS == "darwin" {
		out, err := exec.Command("sw_vers", "-productVersion").Output()
		if v := strings.TrimSpace(string(out)); err == nil && v != "" {
			return "macOS " + v
		}
		return "macOS"
	}
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return "unknown"
	}
	defer f.Close()
	return ParseOSRelease(f)
}

// ParseOSRelease returns PRETTY_NAME, else NAME, else "unknown".
func ParseOSRelease(r io.Reader) string {
	values := map[string]string{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		k, v, ok := strings.Cut(strings.TrimSpace(sc.Text()), "=")
		if !ok {
			continue
		}
		values[k] = strings.Trim(v, `"'`)
	}
	if v := values["PRETTY_NAME"]; v != "" {
		return v
	}
	if v := values["NAME"]; v != "" {
		return v
	}
	return "unknown"
}

func detectFlavor() string {
	out, err := exec.Command("sed", "--version").CombinedOutput()
	resolved := ""
	if p, lerr := exec.LookPath("sed"); lerr == nil {
		if r, rerr := filepath.EvalSymlinks(p); rerr == nil {
			resolved = r
		}
	}
	return Flavor(err == nil, string(out), resolved)
}

// Flavor decides the coreutils flavor from `sed --version` and the resolved sed path.
func Flavor(sedOK bool, sedOutput, sedPath string) string {
	if sedOK && strings.Contains(sedOutput, "GNU") {
		return "GNU"
	}
	if strings.Contains(filepath.Base(sedPath), "busybox") {
		return "BusyBox"
	}
	return "BSD"
}
