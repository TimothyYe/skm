package main

import (
	"runtime/debug"
	"strings"

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

// version returns the go-install module tag, else the Version var (ldflags/default).
func version() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		v := strings.TrimPrefix(info.Main.Version, "v")
		// accept only real release tags, skip devel/dirty/pseudo-versions
		if v != "" && v != "(devel)" && !strings.Contains(v, "+") && !strings.HasPrefix(v, "0.0.0-") {
			return v
		}
	}
	return Version
}

func displayLogo() {
	color.Cyan(logo, version())
}
