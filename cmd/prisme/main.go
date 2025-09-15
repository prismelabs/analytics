package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	var cmd string
	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") {
		cmd = os.Args[1]
		os.Args = os.Args[1:]
	}
	switch cmd {
	case "grafana-dashboard":
		grafanaDashboard()

	case "default-config":
		defaultConfig()

	case "serve":
		fallthrough
	default:
		if cmd != "" && !strings.HasPrefix(cmd, "-") {
			cliError(fmt.Errorf("unknown command %q", cmd))
		}
		serve()
	}
}

func cliError(err error) {
	_, _ = fmt.Fprintln(os.Stderr, "Error:", err.Error())
	_, _ = fmt.Fprintln(os.Stderr)
	_, _ = fmt.Fprintln(os.Stderr, "Usage: prisme [COMMAND] [FLAGS]")
	_, _ = fmt.Fprintln(os.Stderr)
	_, _ = fmt.Fprintf(os.Stderr, "Run 'prisme -h' for more information.\n")
	os.Exit(1)
}
