![](https://raw.githubusercontent.com/TimothyYe/skm/master/snapshots/skm.png)

[![Release][3]][4] [![MIT licensed][5]][6] [![Build Status][1]][2] [![Go Report Card][7]][8] [![GoCover.io][11]][12] [![GoDoc][9]][10]

[1]: https://travis-ci.org/TimothyYe/skm.svg?branch=master
[2]: https://travis-ci.org/TimothyYe/skm
[3]: https://img.shields.io/badge/release-v0.3.2-brightgreen.svg
[4]: https://github.com/TimothyYe/skm/releases
[5]: https://img.shields.io/dub/l/vibe-d.svg
[6]: LICENSE
[7]: https://goreportcard.com/badge/github.com/timothyye/skm
[8]: https://goreportcard.com/report/github.com/timothyye/skm
[9]: https://godoc.org/github.com/TimothyYe/skm?status.svg
[10]: https://godoc.org/github.com/TimothyYe/skm
[11]: https://img.shields.io/badge/gocover.io-81.8%25-green.svg
[12]: https://gocover.io/github.com/timothyye/skm

SKM is a simple and powerful SSH Keys Manager. It helps you to manage your multiple SSH keys easily!

![](https://github.com/TimothyYe/skm/blob/master/snapshots/demo.gif?raw=true)

## Features

* Create, List, Delete your SSH key(s)
* Manage all your SSH keys by alias names
* Choose and set a default SSH key
* Rename SSH key alias name
* Backup and restore all your SSH keys

## Installation

#### Homebrew

```bash
brew tap timothyye/tap
brew install timothyye/tap/skm
```

#### Using Go

```bash
go get github.com/TimothyYe/skm/cmd/skm
```

#### Manual Installation

Download it from [releases](https://github.com/TimothyYe/skm/releases) and extact it to /usr/bin or your PATH directory.

## Usage
```bash
% skm

SKM V0.3.2
https://github.com/TimothyYe/skm

NAME:
   SKM - Manage your multiple SSH keys easily

USAGE:
   skm [global options] command [command options] [arguments...]

VERSION:
   0.3.2

COMMANDS:
     init, i     Initialize SSH keys store for the first time usage.
     create, c   Create a new SSH key.
     ls, l       List all the available SSH keys
     use, u      Set specific SSH key as default by its alias name
     delete, d   Delete specific SSH key by alias name
     rename, rn  Rename SSH key alias name to a new one
     copy, cp    Copy current SSH public key to a remote host
     backup, b   Backup all SSH keys to an archive file
     restore, r  Restore SSH keys from an existing archive file
     help, h     Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
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
__NOTE:__ Currently __ONLY__ RSA key type is supported!

```bash
skm create prod -C "abc@abc.com"

Generating public/private rsa key pair.
Enter passphrase (empty for no passphrase):
Enter same passphrase again:
Your identification has been saved in /Users/timothy/.skm/prod/id_rsa.
Your public key has been saved in /Users/timothy/.skm/prod/id_rsa.pub.
...
✔ SSH key [prod] created!
```

### List SSH keys
```bash
% skm ls

✔ Found 3 SSH key(s)!

->      default
        dev
        prod
```
### Set default SSH key
```bash
% skm use dev
Now using SSH key: dev
```
### Delete a SSH key

```bash
% skm delete prod

Please confirm to delete SSH key [prod] [y/n]: y
✔ SSH key [prod] deleted!
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

## Licence

[MIT License](https://github.com/TimothyYe/skm/blob/master/LICENSE)
