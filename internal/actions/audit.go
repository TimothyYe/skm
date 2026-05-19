package actions

import (
	"bufio"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/TimothyYe/skm/internal/models"
	"github.com/TimothyYe/skm/internal/utils"
	"github.com/fatih/color"
	cli "gopkg.in/urfave/cli.v1"
)

type auditFinding struct {
	Alias   string     `json:"alias"`
	Name    string     `json:"name"`
	Level   checkLevel `json:"level"`
	Message string     `json:"message"`
	Hint    string     `json:"hint,omitempty"`
}

// Audit walks every key in the store and reports hygiene findings: weak RSA
// sizes, unencrypted private keys, and keys older than --max-age. Where doctor
// inspects the environment, audit inspects the keys themselves.
func Audit(c *cli.Context) error {
	env := utils.MustGetEnvironment(c)
	keyMap := utils.LoadSSHKeys(env)

	rsaMin := c.Int("rsa-min")
	if rsaMin <= 0 {
		rsaMin = 3072
	}
	maxAge, err := parseAgeFlag(c.String("max-age"), 365*24*time.Hour)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("invalid --max-age: %v", err), 2)
	}
	strict := c.Bool("strict")

	findings := []auditFinding{}
	aliases := make([]string, 0, len(keyMap))
	for a := range keyMap {
		aliases = append(aliases, a)
	}
	sort.Strings(aliases)

	for _, alias := range aliases {
		key := keyMap[alias]
		findings = append(findings, auditStrength(alias, key, rsaMin)...)
		findings = append(findings, auditPassphrase(alias, key)...)
		findings = append(findings, auditAge(alias, key, maxAge)...)
	}

	if strict {
		for i := range findings {
			if findings[i].Level == levelWarn {
				findings[i].Level = levelFail
			}
		}
	}

	if c.Bool("json") {
		return printAuditJSON(findings)
	}
	printAuditText(findings, len(keyMap))

	for _, f := range findings {
		if f.Level == levelFail {
			return cli.NewExitError("", 1)
		}
	}
	return nil
}

func auditStrength(alias string, key *models.SSHKey, rsaMin int) []auditFinding {
	details := inspectKey(key.PublicKey)
	if details.Fingerprint == "" {
		return []auditFinding{{
			Alias:   alias,
			Name:    "strength",
			Level:   levelWarn,
			Message: "could not inspect public key (ssh-keygen unavailable or key malformed)",
		}}
	}
	keyType := ""
	if key.Type != nil {
		keyType = key.Type.Name
	}
	if keyType != "rsa" {
		return nil
	}
	bits, _ := strconv.Atoi(details.Bits)
	if bits > 0 && bits < rsaMin {
		return []auditFinding{{
			Alias:   alias,
			Name:    "strength",
			Level:   levelFail,
			Message: fmt.Sprintf("RSA-%d is below the %d-bit minimum", bits, rsaMin),
			Hint:    fmt.Sprintf("Rotate with `skm create %s -t ed25519` (or RSA >= %d)", alias, rsaMin),
		}}
	}
	return nil
}

func auditPassphrase(alias string, key *models.SSHKey) []auditFinding {
	encrypted, err := privateKeyEncrypted(key.PrivateKey)
	if err != nil {
		return []auditFinding{{
			Alias:   alias,
			Name:    "passphrase",
			Level:   levelWarn,
			Message: fmt.Sprintf("could not determine passphrase state: %v", err),
		}}
	}
	if encrypted {
		return nil
	}
	return []auditFinding{{
		Alias:   alias,
		Name:    "passphrase",
		Level:   levelWarn,
		Message: "private key is not protected by a passphrase",
		Hint:    fmt.Sprintf("Add one with `skm passphrase %s`", alias),
	}}
}

func auditAge(alias string, key *models.SSHKey, maxAge time.Duration) []auditFinding {
	// Public key mtime is the better proxy for key birth: rotating a passphrase
	// rewrites the private file, but the public file is rewritten only at
	// creation.
	info, err := os.Stat(key.PublicKey)
	if err != nil {
		return nil
	}
	age := time.Since(info.ModTime())
	if age < maxAge {
		return nil
	}
	return []auditFinding{{
		Alias:   alias,
		Name:    "age",
		Level:   levelWarn,
		Message: fmt.Sprintf("key is %s old (threshold %s)", roundDays(age), roundDays(maxAge)),
		Hint:    "Consider rotating with `skm create` + `skm copy`",
	}}
}

// privateKeyEncrypted reports whether a private key file is passphrase-protected.
// Handles both the legacy PEM format (Proc-Type: 4,ENCRYPTED) and the modern
// OpenSSH key format (cipher field != "none").
func privateKeyEncrypted(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var beginLine string
	var bodyLines []string
	inBody := false
	for scanner.Scan() {
		line := scanner.Text()
		if !inBody {
			if strings.HasPrefix(line, "-----BEGIN ") {
				beginLine = line
				inBody = true
				continue
			}
			continue
		}
		if strings.HasPrefix(line, "-----END ") {
			break
		}
		// Legacy PEM headers appear before the blank line that precedes the body.
		if strings.HasPrefix(line, "Proc-Type:") && strings.Contains(line, "ENCRYPTED") {
			return true, nil
		}
		if strings.Contains(line, ":") && !looksLikeBase64(line) {
			continue
		}
		bodyLines = append(bodyLines, strings.TrimSpace(line))
	}
	if err := scanner.Err(); err != nil {
		return false, err
	}
	if beginLine == "" {
		return false, fmt.Errorf("no PEM header found")
	}

	// Legacy formats without Proc-Type are unencrypted.
	if !strings.Contains(beginLine, "OPENSSH PRIVATE KEY") {
		return false, nil
	}
	raw, err := base64.StdEncoding.DecodeString(strings.Join(bodyLines, ""))
	if err != nil {
		return false, fmt.Errorf("base64 decode: %w", err)
	}
	cipher, err := openSSHCipherName(raw)
	if err != nil {
		return false, err
	}
	return cipher != "none", nil
}

// openSSHCipherName extracts the cipher field from the binary OpenSSH private
// key blob. Layout: "openssh-key-v1\0" magic, then length-prefixed strings
// ciphername, kdfname, kdfoptions, ...
func openSSHCipherName(raw []byte) (string, error) {
	const magic = "openssh-key-v1\x00"
	if !strings.HasPrefix(string(raw), magic) {
		return "", fmt.Errorf("missing openssh-key-v1 magic")
	}
	buf := raw[len(magic):]
	if len(buf) < 4 {
		return "", fmt.Errorf("truncated key blob")
	}
	n := binary.BigEndian.Uint32(buf[:4])
	buf = buf[4:]
	if uint32(len(buf)) < n {
		return "", fmt.Errorf("truncated cipher name")
	}
	return string(buf[:n]), nil
}

func looksLikeBase64(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '+', r == '/', r == '=':
		default:
			return false
		}
	}
	return true
}

// parseAgeFlag accepts strings like "30d", "6m", "1y", "12w", or a bare integer
// (interpreted as days). Returns the fallback when value is empty.
func parseAgeFlag(value string, fallback time.Duration) (time.Duration, error) {
	v := strings.TrimSpace(value)
	if v == "" {
		return fallback, nil
	}
	unit := time.Duration(0)
	digits := v
	switch v[len(v)-1] {
	case 'd':
		unit = 24 * time.Hour
		digits = v[:len(v)-1]
	case 'w':
		unit = 7 * 24 * time.Hour
		digits = v[:len(v)-1]
	case 'm':
		unit = 30 * 24 * time.Hour
		digits = v[:len(v)-1]
	case 'y':
		unit = 365 * 24 * time.Hour
		digits = v[:len(v)-1]
	default:
		unit = 24 * time.Hour
	}
	n, err := strconv.Atoi(digits)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("expected <N>[d|w|m|y], got %q", value)
	}
	return time.Duration(n) * unit, nil
}

func roundDays(d time.Duration) string {
	days := int(d.Hours() / 24)
	switch {
	case days >= 365:
		return fmt.Sprintf("%dy", days/365)
	case days >= 30:
		return fmt.Sprintf("%dmo", days/30)
	default:
		return fmt.Sprintf("%dd", days)
	}
}

func printAuditText(findings []auditFinding, totalKeys int) {
	fmt.Println()
	if len(findings) == 0 {
		color.Green("  %s%d key(s) audited, no issues found", utils.CheckSymbol, totalKeys)
		return
	}

	// Group by alias for readability.
	byAlias := map[string][]auditFinding{}
	order := []string{}
	for _, f := range findings {
		if _, ok := byAlias[f.Alias]; !ok {
			order = append(order, f.Alias)
		}
		byAlias[f.Alias] = append(byAlias[f.Alias], f)
	}

	var okN, warnN, failN int
	for _, alias := range order {
		fmt.Printf("[%s]\n", alias)
		for _, f := range byAlias[alias] {
			switch f.Level {
			case levelWarn:
				color.Yellow("  ! %s: %s", f.Name, f.Message)
				if f.Hint != "" {
					color.Yellow("      hint: %s", f.Hint)
				}
				warnN++
			case levelFail:
				color.Red("  %s%s: %s", utils.CrossSymbol, f.Name, f.Message)
				if f.Hint != "" {
					color.Red("      hint: %s", f.Hint)
				}
				failN++
			default:
				okN++
			}
		}
	}

	fmt.Println()
	var failAliases, warnOnlyAliases int
	for _, fs := range byAlias {
		hasFail := false
		hasWarn := false
		for _, f := range fs {
			if f.Level == levelFail {
				hasFail = true
			}
			if f.Level == levelWarn {
				hasWarn = true
			}
		}
		if hasFail {
			failAliases++
		} else if hasWarn {
			warnOnlyAliases++
		}
	}
	clean := totalKeys - len(byAlias)
	summary := fmt.Sprintf("%d clean, %d with warnings, %d with failures (of %d key(s))",
		clean, warnOnlyAliases, failAliases, totalKeys)
	switch {
	case failN > 0:
		color.Red(summary)
	case warnN > 0:
		color.Yellow(summary)
	default:
		color.Green(summary)
	}
}

func printAuditJSON(findings []auditFinding) error {
	out, err := json.MarshalIndent(findings, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	for _, f := range findings {
		if f.Level == levelFail {
			return cli.NewExitError("", 1)
		}
	}
	return nil
}
