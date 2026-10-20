package main

import (
	"fmt"
	"os"
	"sinet/internal/cli"
	"sinet/internal/client"
)

func main() {
	c, err := client.New()
	if err == nil {
		err = cli.Run(c, os.Args[1:], os.Stdin, os.Stdout)
	}
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, cli.SafeText(err.Error()))
		os.Exit(1)
	}
}
