package actions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/TimothyYe/skm/internal/models"
	"github.com/TimothyYe/skm/internal/utils"
	"github.com/fatih/color"
	"gopkg.in/urfave/cli.v1"
)

type keyRow struct {
	Alias       string `json:"alias"`
	Default     bool   `json:"default"`
	Type        string `json:"type"`
	Bits        string `json:"bits,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
	Comment     string `json:"comment,omitempty"`
	InAgent     bool   `json:"in_agent"`
	Created     string `json:"created,omitempty"`
	PrivateKey  string `json:"private_key"`
	PublicKey   string `json:"public_key"`
}

func List(c *cli.Context) error {
	env := utils.MustGetEnvironment(c)
	keyMap := utils.LoadSSHKeys(env)

	if len(keyMap) == 0 {
		color.Green("%s No SSH key found!", utils.CheckSymbol)
		return nil
	}

	aliases := make([]string, 0, len(keyMap))
	for k := range keyMap {
		aliases = append(aliases, k)
	}
	sort.Strings(aliases)

	filterType := strings.ToLower(strings.TrimSpace(c.String("type")))
	agentFps := loadAgentFingerprints()

	rows := make([]keyRow, 0, len(aliases))
	for _, alias := range aliases {
		key := keyMap[alias]
		row := buildRow(alias, key, agentFps)
		if filterType != "" && row.Type != filterType {
			continue
		}
		rows = append(rows, row)
	}

	if c.Bool("json") {
		return printJSON(rows)
	}

	if c.Bool("quiet") {
		printQuiet(rows)
		return nil
	}

	printTable(rows)
	return nil
}

func buildRow(alias string, key *models.SSHKey, agentFps map[string]bool) keyRow {
	row := keyRow{
		Alias:      alias,
		Default:    key.IsDefault,
		PrivateKey: key.PrivateKey,
		PublicKey:  key.PublicKey,
	}
	if key.Type != nil {
		row.Type = key.Type.Name
	}

	if details := inspectKey(key.PublicKey); details.Fingerprint != "" {
		row.Bits = details.Bits
		row.Fingerprint = details.Fingerprint
		row.Comment = details.Comment
		if agentFps[details.Fingerprint] {
			row.InAgent = true
		}
	}

	if info, err := os.Stat(key.PrivateKey); err == nil {
		row.Created = info.ModTime().Format("2006-01-02")
	}
	return row
}

func printTable(rows []keyRow) {
	color.Green("\r\n%sFound %d SSH key(s)!", utils.CheckSymbol, len(rows))
	fmt.Println()

	// Render uncolored through tabwriter first; ANSI escape codes would
	// otherwise be counted as visible width and skew alignment.
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "  ALIAS\tTYPE\tBITS\tFINGERPRINT\tAGENT\tCREATED\tCOMMENT")
	for _, r := range rows {
		marker := "  "
		if r.Default {
			marker = "* "
		}
		agent := ""
		if r.InAgent {
			agent = "yes"
		}
		fmt.Fprintf(w, "%s%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			marker, r.Alias, dash(r.Type), dash(r.Bits), dash(r.Fingerprint), dash(agent), dash(r.Created), dash(r.Comment))
	}
	w.Flush()

	green := color.New(color.FgGreen)
	for line := range strings.SplitSeq(strings.TrimRight(buf.String(), "\n"), "\n") {
		if strings.HasPrefix(line, "* ") {
			green.Println(line)
		} else {
			fmt.Println(line)
		}
	}
}

func printQuiet(rows []keyRow) {
	for _, r := range rows {
		if r.Default {
			color.Green("->\t%s", r.Alias)
		} else {
			color.Blue("\t%s", r.Alias)
		}
	}
}

func printJSON(rows []keyRow) error {
	out, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

func dash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

type keyDetails struct {
	Bits        string
	Fingerprint string
	Comment     string
}

// inspectKey shells out to ssh-keygen -lf to extract bits, fingerprint, and
// comment from a public key. Returns zero values if ssh-keygen is unavailable
// or the file isn't a valid SSH public key.
func inspectKey(pubKeyPath string) keyDetails {
	if pubKeyPath == "" {
		return keyDetails{}
	}
	cmd := exec.Command("ssh-keygen", "-lf", pubKeyPath)
	cmd.Env = append(os.Environ(), "LC_ALL=C")
	out, err := cmd.Output()
	if err != nil {
		return keyDetails{}
	}
	// Format: "<bits> SHA256:<hash> <comment...> (<TYPE>)"
	parts := strings.Fields(strings.TrimSpace(string(out)))
	if len(parts) < 3 {
		return keyDetails{}
	}
	d := keyDetails{Bits: parts[0], Fingerprint: parts[1]}
	if len(parts) > 3 {
		d.Comment = strings.Join(parts[2:len(parts)-1], " ")
	}
	return d
}

// loadAgentFingerprints queries ssh-add -l and returns the set of fingerprints
// currently loaded in the agent. Returns an empty map if the agent isn't
// running or has no identities.
func loadAgentFingerprints() map[string]bool {
	cmd := exec.Command("ssh-add", "-l")
	cmd.Env = append(os.Environ(), "LC_ALL=C")
	// ssh-add exits 1 when there are no identities; combined output handles both.
	out, _ := cmd.Output()
	set := map[string]bool{}
	for line := range strings.SplitSeq(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && strings.HasPrefix(fields[1], "SHA256:") {
			set[fields[1]] = true
		}
	}
	return set
}

