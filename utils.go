package main

import "os"

const (
	Name  = "SKM"
	Usage = "Manage your multiple SSH keys easily"

	CheckSymbol = "\u2714 "
	CrossSymbol = "\u2716 "
)

func parseArgs() {
	if len(os.Args) == 1 {
		displayLogo()
	} else if len(os.Args) == 2 {
		if os.Args[1] == "-h" || os.Args[1] == "--help" || os.Args[1] == "h" || os.Args[1] == "help" {
			displayLogo()
		}
	}
}
