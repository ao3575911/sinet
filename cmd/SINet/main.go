package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sinet"
	"sinet/internal/cli"
	"sinet/internal/client"
)

func main() {
	if len(os.Args) == 2 && (os.Args[1] == "version" || os.Args[1] == "--version") {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]string{"name": "SINet", "version": sinet.Version()})
		return
	}
	var c *client.Client
	var err error
	if len(os.Args) < 2 || (os.Args[1] != "ui" && os.Args[1] != "doctor") {
		c, err = client.New()
	}
	if err == nil {
		err = cli.Run(c, os.Args[1:], os.Stdin, os.Stdout)
	}
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, cli.SafeText(err.Error()))
		os.Exit(1)
	}
}
