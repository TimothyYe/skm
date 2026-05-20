package actions

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/TimothyYe/skm/internal/models"
	"github.com/TimothyYe/skm/internal/utils"
	"github.com/fatih/color"
	"gopkg.in/urfave/cli.v1"
)

func Copy(c *cli.Context) error {
	env := utils.MustGetEnvironment(c)
	host := c.Args().Get(0)
	if host == "" {
		color.Red("%sUsage: skm cp [--key alias] [-p port] [user@]host", utils.CrossSymbol)
		return nil
	}

	host, extractedPort, err := splitHostPort(host)
	if err != nil {
		color.Red("%s%s", utils.CrossSymbol, err.Error())
		return nil
	}
	port := c.String("p")
	if extractedPort != "" {
		if port != "" && port != extractedPort {
			color.Red("%sport conflict: -p %s vs %s embedded in host", utils.CrossSymbol, port, extractedPort)
			return nil
		}
		port = extractedPort
	}

	alias, keyPath, err := resolveKey(c, env)
	if err != nil {
		color.Red("%s%s", utils.CrossSymbol, err.Error())
		return nil
	}

	args := []string{}
	if port != "" {
		args = append(args, "-p", port)
	}
	args = append(args, "-i", keyPath, host)

	if c.Bool("dry-run") {
		color.Yellow("Would run: ssh-copy-id %s", strings.Join(args, " "))
		return nil
	}

	if utils.Execute("", "ssh-copy-id", args...) {
		color.Green("%s SSH key copied to remote host", utils.CheckSymbol)
		_ = utils.RunHook(utils.EventPostCopy, alias, env,
			"SKM_REMOTE_HOST", host,
			"SKM_REMOTE_PORT", port,
		)
	}
	return nil
}

func resolveKey(c *cli.Context, env *models.Environment) (string, string, error) {
	keys := utils.LoadSSHKeys(env)

	if alias := c.String("key"); alias != "" {
		key, ok := keys[alias]
		if !ok {
			return "", "", fmt.Errorf("SSH key [%s] not found", alias)
		}
		return alias, key.PrivateKey, nil
	}

	if c.Bool("pick") {
		alias, err := pickKey("Select an SSH key to copy", keys)
		if err != nil {
			return "", "", err
		}
		return alias, keys[alias].PrivateKey, nil
	}

	// Find the active default by looking for the SSHKey that LoadSSHKeys
	// marked as IsDefault. This correctly handles any supported key type,
	// unlike a hard-coded id_rsa lookup.
	for alias, key := range keys {
		if key.IsDefault {
			return alias, key.PrivateKey, nil
		}
	}
	return "", "", errors.New("No active SSH key found. Run `skm use <alias>` first or pass --key <alias>.")
}

// splitHostPort handles bracketed IPv6 forms like [::1]:2222 and user@[::1]:2222,
// extracting the embedded port and returning a host stripped of brackets. Bare
// IPv6 without a port (e.g. ::1, user@::1) is left unchanged. Hostnames and
// IPv4 with an embedded :port are also split. ssh-copy-id requires the port via
// -p rather than as part of the host argument.
func splitHostPort(host string) (string, string, error) {
	userPrefix := ""
	if at := strings.LastIndex(host, "@"); at >= 0 {
		userPrefix = host[:at+1]
		host = host[at+1:]
	}

	if strings.HasPrefix(host, "[") {
		end := strings.Index(host, "]")
		if end < 0 {
			return "", "", fmt.Errorf("malformed bracketed host: %s", userPrefix+host)
		}
		ip := host[1:end]
		rest := host[end+1:]
		if rest == "" {
			return userPrefix + ip, "", nil
		}
		if !strings.HasPrefix(rest, ":") {
			return "", "", fmt.Errorf("expected :port after bracketed host: %s", userPrefix+host)
		}
		return userPrefix + ip, rest[1:], nil
	}

	// Single colon → host:port (hostnames and IPv4 cannot contain colons).
	// Bare IPv6 has multiple colons and is left intact.
	if strings.Count(host, ":") == 1 {
		if h, p, err := net.SplitHostPort(host); err == nil {
			return userPrefix + h, p, nil
		}
	}
	return userPrefix + host, "", nil
}
