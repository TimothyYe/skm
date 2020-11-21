package main

import (
	"github.com/fatih/color"
)

var (
	// Version is the default version of SKM
	Version = "0.8.1"
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

func displayLogo() {
	color.Cyan(logo, Version)
}
