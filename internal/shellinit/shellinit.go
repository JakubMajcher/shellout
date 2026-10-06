// SPDX-License-Identifier: GPL-3.0-or-later

// Package shellinit prints the optional shell integration scripts.
// The function they define runs the approved command in the current shell,
// so cd/export work and the real command lands in history.
package shellinit

import (
	"fmt"
	"path/filepath"
)

const zshScript = `# shellout integration for zsh: eval "$(shellout init zsh)"
shellout() {
  local __so_file __so_status __so_cmd
  __so_file=$(mktemp) || return 1
  SHELLOUT_EMIT_FILE=$__so_file command shellout "$@"
  __so_status=$?
  if [ "$__so_status" -eq 0 ] && [ -s "$__so_file" ]; then
    __so_cmd=$(<"$__so_file")
    command rm -f -- "$__so_file"
    print -s -- "$__so_cmd"
    eval "$__so_cmd"
    return $?
  fi
  command rm -f -- "$__so_file"
  return $__so_status
}
sho() { shellout "$@"; }
`

const bashScript = `# shellout integration for bash: eval "$(shellout init bash)"
shellout() {
  local __so_file __so_status __so_cmd
  __so_file=$(mktemp) || return 1
  SHELLOUT_EMIT_FILE=$__so_file command shellout "$@"
  __so_status=$?
  if [ "$__so_status" -eq 0 ] && [ -s "$__so_file" ]; then
    __so_cmd=$(<"$__so_file")
    command rm -f -- "$__so_file"
    history -s -- "$__so_cmd"
    eval "$__so_cmd"
    return $?
  fi
  command rm -f -- "$__so_file"
  return $__so_status
}
sho() { shellout "$@"; }
`

const fishScript = `# shellout integration for fish: shellout init fish | source
function shellout
    set -l __so_file (mktemp); or return 1
    SHELLOUT_EMIT_FILE=$__so_file command shellout $argv
    set -l __so_status $status
    if test $__so_status -eq 0; and test -s $__so_file
        set -l __so_cmd (string collect < $__so_file)
        command rm -f -- $__so_file
        if test (string split . -- $version)[1] -ge 4
            history append -- $__so_cmd
        end
        eval $__so_cmd
        return $status
    end
    command rm -f -- $__so_file
    return $__so_status
end
function sho
    shellout $argv
end
`

// Script returns the integration script for zsh, bash or fish.
func Script(shell string) (string, error) {
	switch shell {
	case "zsh":
		return zshScript, nil
	case "bash":
		return bashScript, nil
	case "fish":
		return fishScript, nil
	default:
		return "", fmt.Errorf("unsupported shell %q (supported: zsh, bash, fish)", shell)
	}
}

// RCFile returns the startup file to add the integration to, so the first run
// can print a ready to paste line instead of making the user find it.
func RCFile(shell, home string) (string, error) {
	switch shell {
	case "zsh":
		return filepath.Join(home, ".zshrc"), nil
	case "bash":
		return filepath.Join(home, ".bashrc"), nil
	case "fish":
		return filepath.Join(home, ".config", "fish", "config.fish"), nil
	default:
		return "", fmt.Errorf("no known startup file for shell %q (supported: zsh, bash, fish)", shell)
	}
}
