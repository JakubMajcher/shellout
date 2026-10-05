// SPDX-License-Identifier: GPL-3.0-or-later

// Package app holds the program name and the build version.
package app

// Name is the program name used in messages, paths and environment variables.
const Name = "shellout"

// Version is set at build time with
// -ldflags "-X github.com/JakubMajcher/shellout/internal/app.Version=1.2.3".
var Version = "dev"
