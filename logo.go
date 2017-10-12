package main

import (
	"os"

	"github.com/fatih/color"
)

var (
	Version = "0.1"
	logo    = `
███████╗██╗  ██╗███╗   ███╗
██╔════╝██║ ██╔╝████╗ ████║
███████╗█████╔╝ ██╔████╔██║
╚════██║██╔═██╗ ██║╚██╔╝██║
███████║██║  ██╗██║ ╚═╝ ██║
╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝

SKM V%s
https://github.com/TimothyYe/skm

`
)

func parseArgs() {
	if len(os.Args) == 1 {
		displayLogo()
	} else if len(os.Args) == 2 {
		if os.Args[1] == "-h" || os.Args[1] == "--help" {
			displayLogo()
		}
	}
}

func displayLogo() {
	color.Cyan(logo, Version)
}
