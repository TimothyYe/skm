package actions

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/TimothyYe/skm/internal/publishers"
	"github.com/TimothyYe/skm/internal/utils"
	"github.com/fatih/color"
	cli "gopkg.in/urfave/cli.v1"
)

// Publish uploads an SSH public key to GitHub, GitLab, or Bitbucket via their
// respective APIs. The active key is used when no alias is given. Duplicate
// uploads are detected by canonical key (type + base64) and silently no-op.
func Publish(c *cli.Context) error {
	provider, err := pickProvider(c)
	if err != nil {
		color.Red("%s%s", utils.CrossSymbol, err.Error())
		return nil
	}

	env := utils.MustGetEnvironment(c)
	keys := utils.LoadSSHKeys(env)
	key, alias, err := resolveAliasOrDefault(c, keys)
	if err != nil {
		color.Red("%s%s", utils.CrossSymbol, err.Error())
		return nil
	}

	pubKeyBytes, err := os.ReadFile(key.PublicKey)
	if err != nil {
		color.Red("%sfailed to read public key: %s", utils.CrossSymbol, err.Error())
		return nil
	}
	pubKey := strings.TrimSpace(string(pubKeyBytes))

	extra := map[string]string{}
	if u := c.String("user"); u != "" {
		extra["user"] = u
	}
	p, err := publishers.Resolve(provider, c.String("url"), extra)
	if err != nil {
		color.Red("%s%s", utils.CrossSymbol, err.Error())
		return nil
	}

	token := publishers.ResolveToken(
		c.String("token"),
		tokenEnvVarsFor(provider),
		tokenCLIFor(provider),
	)
	if token == "" {
		color.Red("%s%s", utils.CrossSymbol, publishers.NoTokenError(provider, p.TokenHint()).Error())
		return nil
	}

	title := c.String("title")
	if title == "" {
		title = defaultTitle(alias)
	}

	if c.Bool("dry-run") {
		host := c.String("url")
		if host == "" {
			host = p.Name()
		}
		color.Yellow("Would publish [%s] to %s as %q (key: %s)", alias, host, title, key.PublicKey)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	existingTitle, found, err := p.Existing(ctx, token, pubKey)
	if err != nil {
		color.Red("%scould not check existing keys on %s: %s", utils.CrossSymbol, p.Name(), err.Error())
		return nil
	}
	if found {
		color.Yellow("%s[%s] already published on %s as %q", utils.CheckSymbol, alias, p.Name(), existingTitle)
		return nil
	}

	if err := p.Publish(ctx, token, title, pubKey); err != nil {
		color.Red("%sfailed to publish to %s: %s", utils.CrossSymbol, p.Name(), err.Error())
		return nil
	}
	color.Green("%spublished [%s] to %s as %q", utils.CheckSymbol, alias, p.Name(), title)
	return nil
}

// pickProvider returns the chosen provider name, enforcing that exactly one of
// --github / --gitlab / --bitbucket is set.
func pickProvider(c *cli.Context) (string, error) {
	flags := []string{}
	if c.Bool("github") {
		flags = append(flags, "github")
	}
	if c.Bool("gitlab") {
		flags = append(flags, "gitlab")
	}
	if c.Bool("bitbucket") {
		flags = append(flags, "bitbucket")
	}
	switch len(flags) {
	case 0:
		return "", fmt.Errorf("specify a provider: --github, --gitlab, or --bitbucket")
	case 1:
		return flags[0], nil
	default:
		return "", fmt.Errorf("only one provider flag may be set (got: %s)", strings.Join(flags, ", "))
	}
}

func tokenEnvVarsFor(provider string) []string {
	switch provider {
	case "github":
		return []string{"SKM_GITHUB_TOKEN", "GITHUB_TOKEN", "GH_TOKEN"}
	case "gitlab":
		return []string{"SKM_GITLAB_TOKEN", "GITLAB_TOKEN", "GL_TOKEN"}
	case "bitbucket":
		return []string{"SKM_BITBUCKET_TOKEN", "BITBUCKET_TOKEN"}
	}
	return nil
}

func tokenCLIFor(provider string) []string {
	switch provider {
	case "github":
		return []string{"gh", "auth", "token"}
	case "gitlab":
		return []string{"glab", "auth", "token"}
	}
	// Bitbucket has no widely-installed CLI with a token-print command.
	return nil
}

func defaultTitle(alias string) string {
	host, _ := os.Hostname()
	if host == "" {
		host = "unknown"
	}
	return fmt.Sprintf("skm-%s-%s-%s", alias, host, time.Now().Format("20060102"))
}
