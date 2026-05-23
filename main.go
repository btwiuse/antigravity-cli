package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/anthropics/antigravity-cli/internal/cli/entrypoints"
)

var version = "dev"

func main() {
	app := entrypoints.NewApp(version)

	showVersion := flag.Bool("version", false, "print version and exit")
	showHelp := flag.Bool("help", false, "show help and exit")
	flag.BoolVar(showHelp, "h", false, "show help and exit")
	startTUI := flag.Bool("tui", false, "start the Bubble Tea TUI")
	flag.Usage = func() {
		fmt.Fprintln(os.Stdout, app.Help())
	}
	flag.Parse()

	args := flag.Args()
	switch {
	case *showVersion:
		fmt.Fprintln(os.Stdout, app.Version())
		return
	case *showHelp:
		flag.Usage()
		return
	case *startTUI:
		args = append([]string{"tui"}, args...)
	}

	if err := app.Run(context.Background(), args, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "agy: %v\n", err)
		os.Exit(1)
	}
}
