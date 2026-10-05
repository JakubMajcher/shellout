// SPDX-License-Identifier: GPL-3.0-or-later

// Package llm asks an OpenAI-compatible endpoint for one shell command.
package llm

import (
	"fmt"

	"github.com/JakubMajcher/shellout/internal/sysinfo"
)

const promptTemplate = `You translate a user's request into exactly one shell command.
Environment: OS %s %s, arch %s, shell %s, core utilities %s.
Write the command for that shell and those utilities. If several steps are needed,
join them into one command line with &&, ; or pipes.
"description" is one sentence saying what the command does. If the command deletes or
overwrites data, needs sudo, changes system configuration or sends data over the
network, start the description with "Warning:".`

const standaloneParagraph = `
The command runs in a child process. Changing directory or environment variables has
no effect after it exits, so do not answer with cd, export, source or alias alone.`

// SystemPrompt builds the system message. integration is true when the shell
// function from `shellout init` will run the command in the user's shell.
func SystemPrompt(info sysinfo.Info, integration bool) string {
	p := fmt.Sprintf(promptTemplate, info.OS, info.Distro, info.Arch, info.Shell, info.Flavor)
	if !integration {
		p += standaloneParagraph
	}
	return p
}
