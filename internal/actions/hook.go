package actions

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/TimothyYe/skm/internal/models"
	"github.com/TimothyYe/skm/internal/utils"
	"github.com/fatih/color"
	"gopkg.in/urfave/cli.v1"
)

// HookList prints the hook scripts that would fire for each known event.
//
//   - no args         → global hooks only
//   - alias arg       → global hooks + that alias's per-key hooks
//   - --all           → global hooks + per-key hooks for every alias in the store
func HookList(c *cli.Context) error {
	env := utils.MustGetEnvironment(c)

	if c.Bool("all") {
		return listAll(env)
	}

	var alias string
	if c.NArg() > 0 {
		alias = c.Args().Get(0)
		keys := utils.LoadSSHKeys(env)
		if _, ok := keys[alias]; !ok {
			color.Red("%sKey alias [%s] doesn't exist!", utils.CrossSymbol, alias)
			return nil
		}
	}

	if !printHookSection(alias, env) {
		printEmptyHelp(alias, env)
	}
	return nil
}

func listAll(env *models.Environment) error {
	color.Cyan("== global ==")
	if !printHookSectionGlobalOnly(env) {
		fmt.Println("(no hooks)")
	}

	keys := utils.LoadSSHKeys(env)
	aliases := make([]string, 0, len(keys))
	for k := range keys {
		aliases = append(aliases, k)
	}
	sort.Strings(aliases)

	for _, alias := range aliases {
		fmt.Println()
		color.Cyan("== %s ==", alias)
		if !printHookSectionPerKeyOnly(alias, env) {
			fmt.Println("(no hooks)")
		}
	}
	return nil
}

// printHookSection prints every event with at least one configured hook for
// the given alias (empty alias = global only). Returns true if anything was
// printed.
func printHookSection(alias string, env *models.Environment) bool {
	printed := false
	for _, event := range utils.KnownHookEvents {
		paths := utils.HookPaths(event, alias, env)
		if len(paths) == 0 {
			continue
		}
		printed = true
		color.Cyan("%s", event)
		for _, p := range paths {
			fmt.Printf("  %s  (%s)\n", p, classifyHookPath(p, alias, env))
		}
	}
	return printed
}

// printHookSectionGlobalOnly prints only hooks rooted in the global hooks dir.
func printHookSectionGlobalOnly(env *models.Environment) bool {
	printed := false
	for _, event := range utils.KnownHookEvents {
		p := filepath.Join(env.StorePath, utils.HooksDir, event)
		if !isExecutableHook(p) {
			continue
		}
		printed = true
		color.Cyan("%s", event)
		fmt.Printf("  %s  (global)\n", p)
	}
	return printed
}

// printHookSectionPerKeyOnly prints only per-key hooks for the given alias
// (and the legacy `hook` file, if present).
func printHookSectionPerKeyOnly(alias string, env *models.Environment) bool {
	printed := false
	for _, event := range utils.KnownHookEvents {
		p := filepath.Join(env.StorePath, alias, utils.HooksDir, event)
		if isExecutableHook(p) {
			printed = true
			color.Cyan("%s", event)
			fmt.Printf("  %s  (per-key)\n", p)
		}
	}
	legacy := filepath.Join(env.StorePath, alias, utils.HookName)
	if isExecutableHook(legacy) {
		printed = true
		color.Cyan("%s", utils.EventPostUse)
		fmt.Printf("  %s  (per-key (legacy))\n", legacy)
	}
	return printed
}

func isExecutableHook(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	return info.Mode()&0111 != 0
}

func printEmptyHelp(alias string, env *models.Environment) {
	if alias == "" {
		fmt.Println("No global hooks configured.")
		fmt.Printf("Drop an executable script at %s/<event> to register one.\n",
			filepath.Join(env.StorePath, utils.HooksDir))
	} else {
		fmt.Printf("No hooks configured for [%s] (global or per-key).\n", alias)
	}
	fmt.Printf("Known events: ")
	sorted := append([]string{}, utils.KnownHookEvents...)
	sort.Strings(sorted)
	for i, ev := range sorted {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(ev)
	}
	fmt.Println()
}

func classifyHookPath(path, alias string, env *models.Environment) string {
	globalDir := filepath.Join(env.StorePath, utils.HooksDir)
	if rel, err := filepath.Rel(globalDir, path); err == nil && !startsWithDotDot(rel) {
		return "global"
	}
	if alias != "" {
		legacy := filepath.Join(env.StorePath, alias, utils.HookName)
		if path == legacy {
			return "per-key (legacy)"
		}
		return "per-key"
	}
	return "?"
}

func startsWithDotDot(rel string) bool {
	if rel == ".." {
		return true
	}
	if len(rel) >= 3 && rel[:3] == "../" {
		return true
	}
	if os.PathSeparator != '/' && len(rel) >= 3 && rel[:3] == `..\` {
		return true
	}
	return false
}
