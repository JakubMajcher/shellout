// SPDX-License-Identifier: GPL-3.0-or-later

// Command shellout turns one English sentence into one shell command,
// shows it, asks for confirmation and runs it.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/JakubMajcher/shellout/internal/app"
	"github.com/JakubMajcher/shellout/internal/config"
	"github.com/JakubMajcher/shellout/internal/llm"
	"github.com/JakubMajcher/shellout/internal/runner"
	"github.com/JakubMajcher/shellout/internal/shellinit"
	"github.com/JakubMajcher/shellout/internal/sysinfo"
	"github.com/JakubMajcher/shellout/internal/ui"
)

const usage = `Usage:
  shellout [-p PROFILE] [REQUEST...]   suggest a command, confirm, run it
  shellout init zsh|bash|fish          print shell integration
  shellout -h | --help
  shellout --version

Without REQUEST, shellout reads it from a "> " prompt (no quoting needed).
sho is the same program.
`

type invocation struct {
	profile   string
	help      bool
	version   bool
	isInit    bool
	initShell string
	request   string
}

// parseArgs parses flags up to the first non-flag word. Exactly
// "init <shell>" selects init mode; any other words form the request.
func parseArgs(args []string) (invocation, error) {
	var inv invocation
	fs := flag.NewFlagSet(app.Name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&inv.profile, "p", "", "")
	fs.BoolVar(&inv.help, "h", false, "")
	fs.BoolVar(&inv.help, "help", false, "")
	fs.BoolVar(&inv.version, "version", false, "")
	if err := fs.Parse(args); err != nil {
		return invocation{}, err
	}
	rest := fs.Args()
	if len(rest) == 2 && rest[0] == "init" {
		inv.isInit = true
		inv.initShell = rest[1]
		return inv, nil
	}
	inv.request = strings.Join(rest, " ")
	return inv, nil
}

func main() {
	os.Exit(run(os.Args[1:]))
}

func fail(code int, msg string) int {
	fmt.Fprintf(os.Stderr, "%s: %s\n", app.Name, msg)
	return code
}

func run(args []string) int {
	inv, err := parseArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n%s", app.Name, err, usage)
		return 2
	}
	switch {
	case inv.help:
		fmt.Print(usage)
		return 0
	case inv.version:
		fmt.Println(app.Version)
		return 0
	case inv.isInit:
		s, err := shellinit.Script(inv.initShell)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n%s", app.Name, err, usage)
			return 2
		}
		fmt.Print(s)
		return 0
	}

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return fail(2, "needs an interactive terminal")
	}
	path, err := config.Path(os.Getenv)
	if err != nil {
		return fail(2, err.Error())
	}
	profile, err := config.Load(path, inv.profile, os.Getenv)
	if err != nil {
		return fail(2, err.Error())
	}

	tty := ui.Terminal{In: os.Stdin, Out: os.Stderr}
	request := inv.request
	if request == "" {
		request, err = tty.ReadRequest()
		if err != nil {
			return fail(1, err.Error())
		}
		if request == "" {
			return 1
		}
	}

	info := sysinfo.Collect()
	emitFile := os.Getenv("SHELLOUT_EMIT_FILE")
	client := &llm.Client{
		BaseURL: profile.BaseURL,
		Model:   profile.Model,
		APIKey:  profile.APIKey,
		Profile: profile.Name,
		Params:  profile.Params,
		Timeout: profile.Timeout,
		HTTP:    &http.Client{},
	}

	stop := ui.StartSpinner(os.Stderr, term.IsTerminal(int(os.Stderr.Fd())))
	suggestion, err := client.Suggest(context.Background(), llm.SystemPrompt(info, emitFile != ""), request)
	stop()
	if err != nil {
		return fail(1, err.Error())
	}

	command, ok, err := tty.Confirm(suggestion.Description, suggestion.Command)
	if err != nil {
		return fail(1, err.Error())
	}
	if !ok {
		return 1
	}

	if emitFile != "" {
		if err := os.WriteFile(emitFile, []byte(command), 0o600); err != nil {
			return fail(1, err.Error())
		}
		return 0
	}
	code, err := runner.Run(info.ShellPath, command, profile.CommandTimeout)
	if err != nil {
		if errors.Is(err, runner.ErrTimeout) {
			return fail(code, err.Error())
		}
		return fail(1, fmt.Sprintf("cannot run %s: %v", info.ShellPath, err))
	}
	return code
}
