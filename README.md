![](https://raw.githubusercontent.com/TimothyYe/skm/master/assets/snapshots/skm.png)

[![MIT licensed][5]][6] [![LICENSE](https://img.shields.io/badge/license-NPL%20(The%20996%20Prohibited%20License)-blue.svg)](https://github.com/996icu/996.ICU/blob/master/LICENSE) [![Build Status][1]][2] [![Go Report Card][7]][8] [![Go Reference][9]][10]

[1]: https://github.com/TimothyYe/skm/actions/workflows/go.yml/badge.svg?branch=master
[2]: https://github.com/TimothyYe/skm/actions/workflows/go.yml
[5]: https://img.shields.io/dub/l/vibe-d.svg
[6]: LICENSE
[7]: https://goreportcard.com/badge/github.com/timothyye/skm
[8]: https://goreportcard.com/report/github.com/timothyye/skm
[9]: https://pkg.go.dev/badge/github.com/TimothyYe/skm.svg
[10]: https://pkg.go.dev/github.com/TimothyYe/skm

SKM is a simple and powerful SSH Keys Manager. It helps you to manage your multiple SSH keys easily!

[中文版 README](README_CN.md)

![](https://github.com/TimothyYe/skm/blob/master/assets/snapshots/demo.gif?raw=true)

## Features

* Create, List, Delete your SSH key(s)
* Manage all your SSH keys by alias names
* Choose and set a default SSH key
* Display public key via alias name
* Copy any SSH key to a remote host (with optional `--key`, `--pick`, `--dry-run`)
* Rename SSH key alias name
* Backup and restore all your SSH keys
* Import an existing key pair from anywhere on disk
* Export a single key as a (optionally encrypted) bundle
* Inspect a key with `fingerprint` and `info`
* Add / rotate / remove a key's passphrase
* Diagnose your environment with `skm doctor`
* Audit stored keys for weak strength, missing passphrases, and age with `skm audit`
* Soft-delete to a recoverable trash, with `skm trash list|restore|empty`
* Prompt UI (with fuzzy search) for SSH key selection across multiple commands
* Customized SSH key store path
* Pluggable hooks on `post-use`, `post-create`, `pre-delete`, `post-copy` events (per-key and global)

## Installation

#### Homebrew

Starting from **v0.8.9**, `skm` has been officially submitted to the [homebrew-core](https://github.com/Homebrew/homebrew-core) repository, so you can install it directly on both macOS and Linux:

```bash
brew install skm
```

> If you previously installed `skm` through the old tap, please remove it first to avoid conflicts:
>
> ```bash
> brew uninstall skm
> brew untap timothyye/tap
> brew install skm
> ```

#### Using Go

```bash
go get github.com/TimothyYe/skm/cmd/skm
```

#### Manual Installation

Download it from [releases](https://github.com/TimothyYe/skm/releases) and extact it to /usr/bin or your PATH directory.

## Usage
```bash
% skm

SKM V0.8.8
https://github.com/TimothyYe/skm

NAME:
   SKM - Manage your multiple SSH keys easily

USAGE:
   skm [global options] command [command options] [arguments...]

VERSION:
   0.8.8

COMMANDS:
     init, i          Initialize SSH keys store for the first time use.
     create, c        Create a new SSH key.
     ls, l            List all the available SSH keys.
     use, u           Set specific SSH key as default by its alias name.
     delete, d        Delete specific SSH key by alias name.
     rename, rn       Rename SSH key alias name to a new one.
     copy, cp         Copy SSH public key to a remote host.
     display, dp      Display the current SSH public key or specific SSH public key by alias name.
     backup, b        Backup all SSH keys to an archive file.
     restore, r       Restore SSH keys from an existing archive file.
     import, im       Import an existing SSH key pair (or an skm export bundle) into the store.
     export, ex       Export a single key as a tar.gz bundle (optionally encrypted).
     fingerprint, fp  Print the SHA256 fingerprint of an SSH key.
     info, in         Show detailed information about an SSH key.
     passphrase, pp   Add, rotate, or remove the passphrase on an SSH key.
     doctor, dr       Run diagnostics against the SKM environment, agent, and stored keys.
     audit, au        Audit stored keys for weak strength, missing passphrases, and age.
     cache            Add your SSH to SSH agent cache via alias name.
     help, h          Shows a list of commands or help for one command.

GLOBAL OPTIONS:
   --store-path value   Path where SKM should store its profiles (default: "/Users/timothy/.skm")
   --ssh-path value     Path to a .ssh folder (default: "/Users/timothy/.ssh")
   --restic-path value  Path to the restic binary
   --help, -h           show help
   --version, -v        print the version
```

### For the first time use

You should initialize the SSH key store for the first time use:

```bash
% skm init

✔ SSH key store initialized!
```

So, where are my SSH keys?
SKM will create SSH key store at ```$HOME/.skm``` and put all the SSH keys in it.

__NOTE:__ If you already have id_rsa & id_rsa.pub key pairs in ```$HOME/.ssh```, SKM will move them to ```$HOME/.skm/default```

### Create a new SSH key
__NOTE:__ Currently __ONLY__ RSA and ED25519 keys are supported!

```bash
skm create prod -C "abc@abc.com" -t ed25519

Generating public/private rsa key pair.
Enter passphrase (empty for no passphrase):
Enter same passphrase again:
Your identification has been saved in /Users/timothy/.skm/prod/id_rsa.
Your public key has been saved in /Users/timothy/.skm/prod/id_rsa.pub.
...
✔ SSH key [prod] created!
```

### List SSH keys

By default `ls` shows alias, key type, and comment in three columns; the
active default key is marked with `->`. Use `-l` / `--long` for the full table
with bits, fingerprint, agent status, and the modification date.

```bash
% skm ls

✔ Found 3 SSH key(s)!

->  default  [ssh-ed25519]  [work@laptop]
    dev      [ssh-ed25519]  [dev]
    prod     [ssh-rsa]      [prod]

% skm ls -l

✔ Found 3 SSH key(s)!

  ALIAS    TYPE     BITS  FINGERPRINT                                         AGENT  CREATED     COMMENT
* default  ed25519  256   SHA256:pFsC7J7L9L08f3w8uP6ozRaGW5Dg8CdEkP8iVj7++pw  yes    2026-05-16  work@laptop
  dev      ed25519  256   SHA256:7dFJEj7WGAL8rn9AqLNYdoTQrqgv00kdnqJlufvxgg4  -      2026-05-12  dev
  prod     rsa      4096  SHA256:DEyhI38hQ5WYABdx9SrJuhqrIyLvfRcZTtzXARuyn0k  -      2026-05-01  prod
```

Other useful flags:

```bash
skm ls -q              # quiet — just the alias names, default marked with ->
skm ls --json          # machine-readable output for scripting
skm ls -t ed25519      # filter by key type
```
### Set default SSH key
```bash
% skm use dev
Now using SSH key: dev
```

### Prompt UI for key selection

You can just type ```skm use```, then a prompt UI will help you to choose the right SSH key:

![](https://github.com/TimothyYe/skm/blob/master/assets/snapshots/prompt.gif?raw=true)

### Display public key

```bash
% skm display
```

Or display specific SSH public key by alias name:

```bash
% skm display prod
```

### Delete a SSH key

```bash
% skm delete prod

Please confirm to delete SSH key [prod] [y/n]: y
✔ SSH key [prod] moved to trash (restore with: skm trash restore prod-20260521150412)
```

By default a delete moves the alias into the store's trash so it can be recovered. Pass multiple aliases to batch-delete; missing aliases are reported and skipped. `-y` / `--yes` skips the confirmation, `--purge` hard-deletes (skipping the trash):

```bash
% skm delete -y staging legacy old-laptop
% skm delete --purge --yes ancient
```

### Recover a deleted SSH key

```bash
% skm trash list
NAME                       ALIAS    DELETED
prod-20260521150412        prod     2026-05-21 15:04:12

% skm trash restore prod-20260521150412
✔ Restored [prod-20260521150412] as alias [prod]
```

You can pass either the trash entry name (`prod-20260521150412`) or just the alias (`prod`); the latter works when only one trashed entry matches. If the original alias is already taken, pass `--as <new-alias>`. Empty the trash with `skm trash empty` (prompts for confirmation; `-y` to skip).
### Copy SSH public key to a remote host

By default the currently active key is pushed:

```bash
% skm cp timothy@example.com

/usr/bin/ssh-copy-id: INFO: Source of key(s) to be installed: "/Users/timothy/.skm/default/id_rsa.pub"
...
✔  SSH key copied to remote host
```

Useful flags:

```bash
skm cp --key work timothy@example.com         # push a specific key, not the default
skm cp --pick timothy@example.com             # interactively choose the key
skm cp -p 2222 timothy@example.com            # non-default SSH port
skm cp timothy@[2001:db8::1]:2222             # IPv6 hosts work too
skm cp --dry-run timothy@example.com          # preview the ssh-copy-id command
```

### Rename a SSH key with a new alias name

```bash
% skm rn test tmp
✔  SSH key [test] renamed to [tmp]
```

### Backup SSH keys

Backup all your SSH keys to $HOME directory by default.

```bash
% skm backup

a .
a ./test
a ./default
a ./dev
a ./dev/id_rsa
a ./dev/id_rsa.pub
a ./default/id_rsa
a ./default/id_rsa.pub
a ./test/id_rsa
a ./test/id_rsa.pub

✔  All SSH keys backup to: /Users/timothy/skm-20171016170707.tar
```

If you have [restic](https://restic.net/) installed then you can also use that
to create backups of your SKM store:

```bash
# First, you need a password for your repository
% if [[ ! -f ~/.skm-backups.passwd ]]; then
%     openssl rand -hex 64 > ~/.skm-backups.passwd
% fi

% skm backup --restic
repository ... opened successfully, password is correct

Files:           0 new,     1 changed,     4 unmodified
Dirs:            0 new,     0 changed,     0 unmodified
Added to the repo: 1.179 KiB

processed 5 files, 2.593 KiB in 0:00
snapshot $SNAPSHOT saved
✔  Backup to /Users/$USER/.skm-backups complete
```

### Restore SSH keys

```bash
% skm restore ~/skm-20171016172828.tar.gz                                                                                           
x ./
x ./test/
x ./default/
x ./dev/
x ./dev/id_rsa
x ./dev/id_rsa.pub
x ./default/._id_rsa
x ./default/id_rsa
x ./default/._id_rsa.pub
x ./default/id_rsa.pub
x ./test/id_rsa
x ./test/id_rsa.pub

✔  All SSH keys restored to /Users/timothy/.skm
```

Again, SKM also supports [restic](https://restic.net/) to create and restore
backups:

```bash
% skm restore --restic --restic-snapshot $SNAPSHOT
repository $REPO opened successfully, password is correct
restoring <Snapshot $SNAPSHOT of [/Users/$USER/.skm] at 2018-10-03 19:40:33.333130348 +0200 CEST by $USER@$HOST> to /Users/$USER/.skm
✔  Backup restored to /Users/$USER/.skm
```

### Import an existing key

Pull an existing key pair from anywhere on disk into the SKM store. The
matching half of the pair is inferred from the `.pub` suffix; the key type
is detected from the public key's header.

```bash
% skm import --alias work ~/old-laptop/.ssh/id_ed25519
✔ Imported ed25519 key as [work]
```

`skm import` also accepts an `skm export` bundle (`.tar.gz`, `.tgz`, or
`.tar.gz.enc`). For encrypted bundles, `openssl` prompts for the passphrase.

```bash
% skm import ~/skm-work-20260516193629.tar.gz.enc
enter aes-256-cbc decryption password:
✔ Imported bundle as [work]
```

Useful flags:

```bash
skm import --alias newname bundle.tar.gz      # rename the alias on the way in
skm import --move ~/old/id_ed25519            # remove the source after a successful copy
```

> **Note:** put `--alias` / `--move` *before* the path argument — flags placed
> after the path are not parsed, and SKM will error out with a hint if it
> sees one.

### Export a single key

Bundle one alias into a portable archive that you can move to another
machine and import there. The archive contains the private key, the public
key, and any `hook` file under that alias.

```bash
% skm export work
✔ Exported [work] to /Users/timothy/skm-work-20260516193629.tar.gz
```

Add `--encrypt` to wrap the bundle with `openssl enc -aes-256-cbc -pbkdf2`,
which prompts for a passphrase:

```bash
% skm export --encrypt work
enter aes-256-cbc encryption password:
Verifying - enter aes-256-cbc encryption password:
✔ Exported [work] to /Users/timothy/skm-work-20260516193629.tar.gz.enc
```

You can also specify the output path:

```bash
skm export -o /tmp/work.tar.gz work
```

### Inspect a key

```bash
% skm fingerprint work
SHA256:pFsC7J7L9L08f3w8uP6ozRaGW5Dg8CdEkP8iVj7++pw

% skm info work
Alias        work
Default      yes
Type         ed25519
Bits         256
Fingerprint  SHA256:pFsC7J7L9L08f3w8uP6ozRaGW5Dg8CdEkP8iVj7++pw
Comment      work@laptop
In agent     yes
Private      /Users/timothy/.skm/work/id_ed25519
Public       /Users/timothy/.skm/work/id_ed25519.pub
Modified     2026-05-16 19:56:07
```

Both commands default to the active key when no alias is given.

### Add / rotate / remove a passphrase

`skm passphrase` wraps `ssh-keygen -p`. Use an empty new passphrase at the
prompt to remove the existing one.

```bash
% skm passphrase work
Updating passphrase for [work] (/Users/timothy/.skm/work/id_ed25519)
Enter new passphrase (empty for no passphrase):
Enter same passphrase again:
Your identification has been saved with the new passphrase.
✔ Passphrase updated for [work]
```

### Run diagnostics

`skm doctor` walks through the most common environment problems — missing
SSH binaries, an unreachable `ssh-agent`, unwritable store paths, the
default key not resolving, RSA keys below 3072 bits, public/private files
with loose permissions, or hook scripts missing the executable bit.

```bash
% skm doctor
  ✔ ssh-keygen available
  ✔ ssh-add available
  ✔ ssh-copy-id available
  ✔ ssh-agent reachable with identities loaded
  ✔ /Users/timothy/.skm writable (mode -rwxr-xr-x)
  ✔ /Users/timothy/.ssh writable (mode -rwxr-xr-x)
  ✔ default key [work] resolves to /Users/timothy/.skm/work/id_ed25519
  ✔ private key id_ed25519 mode -rw-------
  ✔ key [work] ed25519/256
  ✔ hook for [work] is executable

10 passed, 0 warning(s), 0 failure(s)
```

`skm doctor --json` emits the same checks as a JSON array for scripting.
Failures cause a non-zero exit; warnings do not.

### Audit your keys

Where `doctor` inspects the environment, `skm audit` (alias `au`) inspects the
keys themselves. It walks every key in the store and reports:

* RSA keys below `--rsa-min` (default 3072) — **failure**
* Private keys with no passphrase — **warning**
* Keys older than `--max-age` (default `1y`, accepts `Nd|Nw|Nm|Ny`) — **warning**

```bash
% skm audit
[legacy]
  ✖ strength: RSA-1024 is below the 3072-bit minimum
      hint: Rotate with `skm create legacy -t ed25519` (or RSA >= 3072)
  ! passphrase: private key is not protected by a passphrase
      hint: Add one with `skm passphrase legacy`
  ! age: key is 2y old (threshold 1y)
      hint: Consider rotating with `skm create` + `skm copy`

1 clean, 0 with warnings, 1 with failures (of 2 key(s))
```

Flags:

* `--json` — emit findings as a JSON array for tooling
* `--strict` — promote warnings to failures (useful in CI)
* `--max-age <duration>` — override the age threshold (e.g. `30d`, `6m`)
* `--rsa-min <bits>` — override the minimum acceptable RSA key size

Audit exits non-zero on any failure-level finding; warnings alone return 0
unless `--strict` is set.

### Integrate with SSH agent

You can use `cache` command to cache your SSH key into SSH agent's cache via SSH alias name.

__Cache your SSH key__  

```bash
λ tim [~/]
→ skm cache --add my                                                                                                                                                                                                                                                                     
Enter passphrase for /Users/timothy/.skm/my/id_rsa:
Identity added: /Users/timothy/.skm/my/id_rsa (/Users/timothy/.skm/my/id_rsa)
✔  SSH key [my] already added into cache
```

__Remove your SSH key from cache__  

```bash
λ tim [~/]
→ ./skm cache --del my                                                                                                                                                                                                                                                                   
Identity removed: /Users/timothy/.skm/my/id_rsa (MyKEY)
✔  SSH key [my] removed from cache
```

__List your cached SSH keys from SSH agent__  

```bash
λ tim [~/]
→ ./skm cache --list                                                                                                                                                                                                                                                                     
2048 SHA256:qAVcwc0tdUOCjH3sTskwxAmfMQiL2sKtfPBXFnUoZHQ /Users/timothy/.skm/my/id_rsa (RSA)
```

### Customized SSH key store path

By default, SKM uses `$HOME/.skm` as the default path of SSH key store.
You can define your customized key store path in your `~/.bashrc` or `~/.zshrc` by adding:

```bash
SKM_STORE_PATH=/usr/local/.skm
```

### Hook mechanism

SKM fires hook scripts around key-lifecycle events. Hooks are ordinary executables (binary or script) that you drop into a known location; SKM picks them up automatically.

#### Events

| Event         | When it fires                              | Failure behavior              |
| ------------- | ------------------------------------------ | ----------------------------- |
| `post-use`    | After switching the default SSH key        | Best-effort (warning only)    |
| `post-create` | After `skm create` succeeds                | Best-effort (warning only)    |
| `pre-delete`  | After you confirm a delete, before removal | **Non-zero exit aborts the delete** |
| `post-copy`   | After `skm copy` (ssh-copy-id) succeeds    | Best-effort (warning only)    |

#### Where to put hook scripts

Two scopes; both fire (global first, then per-key):

```bash
~/.skm/hooks/<event>              # global — runs for any alias
~/.skm/<alias>/hooks/<event>      # per-key — runs only for that alias
```

For example:

```bash
~/.skm/hooks/pre-delete           # global guard for every delete
~/.skm/work/hooks/post-use        # only when switching to "work"
~/.skm/prod/hooks/post-copy       # only after copying the "prod" key
```

The legacy `~/.skm/<alias>/hook` file is still honored as a `post-use` hook for backward compatibility.

Don't forget to make scripts executable:

```bash
chmod +x ~/.skm/work/hooks/post-use
```

#### Environment variables passed to hooks

Every hook receives these in its environment:

| Variable           | Description                                      |
| ------------------ | ------------------------------------------------ |
| `SKM_EVENT`        | The event name (`post-use`, `pre-delete`, …)     |
| `SKM_ALIAS`        | The alias name (also passed as `$1`)             |
| `SKM_STORE_PATH`   | The SKM store path                               |
| `SKM_SSH_PATH`     | The user's `~/.ssh` path                         |
| `SKM_KEY_TYPE`     | `rsa`, `ed25519`, …                              |
| `SKM_PRIVATE_KEY`  | Absolute path to the private key                 |
| `SKM_PUBLIC_KEY`   | Absolute path to the public key                  |

Event-specific extras:

| Event       | Extra variables                          |
| ----------- | ---------------------------------------- |
| `post-copy` | `SKM_REMOTE_HOST`, `SKM_REMOTE_PORT`     |

#### Example: switch git identity on `post-use`

```bash
#!/bin/bash
# ~/.skm/work/hooks/post-use
git config --global user.name  "Your Work Name"
git config --global user.email "you@work.example.com"
```

#### Example: log every key deployment via `post-copy`

```bash
#!/bin/sh
# ~/.skm/hooks/post-copy
echo "$(date -Iseconds) $SKM_ALIAS -> $SKM_REMOTE_HOST:${SKM_REMOTE_PORT:-22}" \
  >> ~/.skm/deployments.log
```

#### Example: guard production keys against accidental delete

```bash
#!/bin/sh
# ~/.skm/prod/hooks/pre-delete — non-zero exit aborts the delete
echo "Refusing to delete production key. Remove this hook first." >&2
exit 1
```

#### Inspect configured hooks

```bash
skm hook ls              # global hooks
skm hook ls <alias>      # global + per-key for one alias
skm hook ls --all        # global + per-key for every alias
```

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=timothyye/skm&type=date&legend=top-left)](https://www.star-history.com/#timothyye/skm&type=date&legend=top-left)

## Licence

[MIT License](https://github.com/TimothyYe/skm/blob/master/LICENSE)  
