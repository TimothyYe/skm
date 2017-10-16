package main

import (
	"github.com/fatih/color"
)

var (
	//Version of SKM
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

func displayLogo() {
	color.Cyan(logo, Version)
}
