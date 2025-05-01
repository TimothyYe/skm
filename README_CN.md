![](https://raw.githubusercontent.com/TimothyYe/skm/master/assets/snapshots/skm.png)

[![MIT licensed][5]][6] [![LICENSE](https://img.shields.io/badge/license-NPL%20(The%20996%20Prohibited%20License)-blue.svg)](https://github.com/996icu/996.ICU/blob/master/LICENSE) [![Build Status][1]][2] [![Go Report Card][7]][8] [![GoCover.io][11]][12] [![GoDoc][9]][10]

[1]: https://travis-ci.org/TimothyYe/skm.svg?branch=master
[2]: https://travis-ci.org/TimothyYe/skm
[5]: https://img.shields.io/dub/l/vibe-d.svg
[6]: LICENSE
[7]: https://goreportcard.com/badge/github.com/timothyye/skm
[8]: https://goreportcard.com/report/github.com/timothyye/skm
[9]: https://godoc.org/github.com/TimothyYe/skm?status.svg
[10]: https://godoc.org/github.com/TimothyYe/skm
[11]: https://img.shields.io/badge/gocover.io-81.8%25-green.svg
[12]: https://gocover.io/github.com/timothyye/skm

SKM 是一个简单而强大的 SSH 密钥管理工具。它帮助您轻松管理多个 SSH 密钥！

![](https://github.com/TimothyYe/skm/blob/master/assets/snapshots/demo.gif?raw=true)

## 功能

* 创建、列出、删除您的 SSH 密钥
* 通过别名管理所有 SSH 密钥
* 选择并设置默认 SSH 密钥
* 通过别名显示公钥
* 将默认 SSH 密钥复制到远程主机
* 重命名 SSH 密钥别名
* 备份和恢复所有 SSH 密钥
* SSH 密钥选择的提示界面
* 自定义 SSH 密钥存储路径

## 安装

#### Homebrew

```bash
brew tap timothyye/tap
brew install timothyye/tap/skm
```

#### 使用 Go

```bash
go get github.com/TimothyYe/skm/cmd/skm
```

#### 手动安装

从 [releases](https://github.com/TimothyYe/skm/releases) 下载并解压到 /usr/bin 或您的 PATH 目录。

## 使用方法
```bash
% skm

SKM V0.8.5
https://github.com/TimothyYe/skm

NAME:
   SKM - Manage your multiple SSH keys easily

USAGE:
   skm [global options] command [command options] [arguments...]

VERSION:
   0.8.5

COMMANDS:
     init, i      Initialize SSH keys store for the first time usage.
     create, c    Create a new SSH key.
     ls, l        List all the available SSH keys.
     use, u       Set specific SSH key as default by its alias name.
     delete, d    Delete specific SSH key by alias name.
     rename, rn   Rename SSH key alias name to a new one.
     copy, cp     Copy current SSH public key to a remote host.
     display, dp  Display the current SSH public key or specific SSH public key by alias name.
     backup, b    Backup all SSH keys to an archive file.
     restore, r   Restore SSH keys from an existing archive file.
     cache        Add your SSH to SSH agent cache via alias name.
     help, h      Shows a list of commands or help for one command.

GLOBAL OPTIONS:
   --store-path value   Path where SKM should store its profiles (default: "/Users/timothy/.skm")
   --ssh-path value     Path to a .ssh folder (default: "/Users/timothy/.ssh")
   --restic-path value  Path to the restic binary
   --help, -h           show help
   --version, -v        print the version
```

### 首次使用

首次使用时，您应该初始化 SSH 密钥存储：

```bash
% skm init

✔ SSH key store initialized!
```

那么，我的 SSH 密钥在哪里？
SKM 将在 ```$HOME/.skm``` 创建 SSH 密钥存储，并将所有 SSH 密钥放入其中。

__注意：__ 如果您已经在 ```$HOME/.ssh``` 中有 id_rsa 和 id_rsa.pub 密钥对，SKM 将它们移动到 ```$HOME/.skm/default```

### 创建新的 SSH 密钥
__注意：__ 目前 __仅__ 支持 RSA 和 ED25519 密钥！

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

### 列出 SSH 密钥
```bash
% skm ls

✔ Found 3 SSH key(s)!

->      default
        dev
        prod
```
### 设置默认 SSH 密钥
```bash
% skm use dev
Now using SSH key: dev
```

### 密钥选择的提示界面

您可以直接输入 ```skm use```，然后一个提示界面将帮助您选择正确的 SSH 密钥：

![](https://github.com/TimothyYe/skm/blob/master/assets/snapshots/prompt.gif?raw=true)

### 显示公钥

```bash
% skm display
```

或通过别名显示特定的 SSH 公钥：

```bash
% skm display prod
```

### 删除 SSH 密钥

```bash
% skm delete prod

Please confirm to delete SSH key [prod] [y/n]: y
✔ SSH key [prod] deleted!
```
### 将 SSH 公钥复制到远程主机

```bash
% skm cp timothy@example.com

/usr/bin/ssh-copy-id: INFO: Source of key(s) to be installed: "/Users/timothy/.skm/default/id_rsa.pub"
/usr/bin/ssh-copy-id: INFO: attempting to log in with the new key(s), to filter out any that are already installed
/usr/bin/ssh-copy-id: INFO: 1 key(s) remain to be installed -- if you are prompted now it is to install the new keys
timothy@example.com's password:

Number of key(s) added:        1

Now try logging into the machine, with:   "ssh 'timothy@example.com'"
and check to make sure that only the key(s) you wanted were added.

✔  Current SSH key already copied to remote host
```

### 重命名 SSH 密钥别名

```bash
% skm rn test tmp
✔  SSH key [test] renamed to [tmp]
```

### 备份 SSH 密钥

默认情况下，将所有 SSH 密钥备份到 $HOME 目录。

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

如果您安装了 [restic](https://restic.net/)，也可以使用它来创建备份：

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

### 恢复 SSH 密钥

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

同样，SKM 也支持 [restic](https://restic.net/) 创建和恢复备份：

```bash
% skm restore --restic --restic-snapshot $SNAPSHOT
repository $REPO opened successfully, password is correct
restoring <Snapshot $SNAPSHOT of [/Users/$USER/.skm] at 2018-10-03 19:40:33.333130348 +0200 CEST by $USER@$HOST> to /Users/$USER/.skm
✔  Backup restored to /Users/$USER/.skm
```

### 与 SSH 代理集成

您可以使用 `cache` 命令通过 SSH 别名将 SSH 密钥缓存到 SSH 代理的缓存中。

__缓存您的 SSH 密钥__  

```bash
λ tim [~/]
→ skm cache --add my                                                                                                                                                                                                                                                                     
Enter passphrase for /Users/timothy/.skm/my/id_rsa:
Identity added: /Users/timothy/.skm/my/id_rsa (/Users/timothy/.skm/my/id_rsa)
✔  SSH key [my] already added into cache
```

__从缓存中删除您的 SSH 密钥__  

```bash
λ tim [~/]
→ ./skm cache --del my                                                                                                                                                                                                                                                                   
Identity removed: /Users/timothy/.skm/my/id_rsa (MyKEY)
✔  SSH key [my] removed from cache
```

__列出 SSH 代理中缓存的 SSH 密钥__  

```bash
λ tim [~/]
→ ./skm cache --list                                                                                                                                                                                                                                                                     
2048 SHA256:qAVcwc0tdUOCjH3sTskwxAmfMQiL2sKtfPBXFnUoZHQ /Users/timothy/.skm/my/id_rsa (RSA)
```

### 自定义 SSH 密钥存储路径

默认情况下，SKM 使用 `$HOME/.skm` 作为 SSH 密钥存储的默认路径。
您可以在 `~/.bashrc` 或 `~/.zshrc` 中定义自定义的密钥存储路径，方法是添加：

```bash
SKM_STORE_PATH=/usr/local/.skm
```

### 钩子机制

在指定的密钥目录中编辑并放置一个名为 ```hook``` 的可执行文件，例如：

```bash
~/.skm/prod/hook
```

这个钩子文件可以是可执行的二进制文件或可执行的脚本文件。

SKM 会在切换默认 SSH 密钥后调用这个钩子文件，您可以在这个钩子文件中做一些操作。

例如，如果您希望在切换到不同的 SSH 密钥后使用不同的 git 用户名和电子邮件，您可以创建一个钩子文件，并在其中放置 shell 命令：

```bash
#!/bin/bash
git config --global user.name "YourNewName"
git config --global user.email "YourNewEmail@example.com"
```

然后使这个钩子文件可执行：

```bash
chmod +x hook
```

SKM 将调用这个钩子文件并为您更改 git 全局设置！

## 许可证

[MIT 许可证](https://github.com/TimothyYe/skm/blob/master/LICENSE)  
