package actions

import (
	"errors"
	"sort"
	"strings"

	"github.com/TimothyYe/skm/internal/models"
	"github.com/manifoldco/promptui"
)

// ErrNoKeys is returned by pickKey when the store is empty.
var ErrNoKeys = errors.New("no SSH keys found in the store")

// pickKey shows an interactive, fuzzy-searchable selection prompt over the
// aliases in keyMap and returns the chosen alias. Used by use/delete/display/
// rename/copy when the user didn't pass an alias on the command line.
func pickKey(label string, keyMap map[string]*models.SSHKey) (string, error) {
	if len(keyMap) == 0 {
		return "", ErrNoKeys
	}

	names := make([]string, 0, len(keyMap))
	for k := range keyMap {
		names = append(names, k)
	}
	sort.Strings(names)

	templates := &promptui.SelectTemplates{
		Active:   "{{ . | white | bgGreen }} ",
		Inactive: "{{ . }} ",
		Selected: "{{ . | bold }} ",
	}

	prompt := promptui.Select{
		Label:             label,
		Items:             names,
		Templates:         templates,
		StartInSearchMode: true,
		Searcher: func(input string, index int) bool {
			return strings.Contains(strings.ToLower(names[index]), strings.ToLower(input))
		},
	}

	_, result, err := prompt.Run()
	if err != nil {
		return "", err
	}
	return result, nil
}

// promptText runs a free-form text prompt. Used to capture the destination
// alias for `skm rename` when only one positional argument was supplied.
func promptText(label string, validate promptui.ValidateFunc) (string, error) {
	p := promptui.Prompt{Label: label, Validate: validate}
	return p.Run()
}
