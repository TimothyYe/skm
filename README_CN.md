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

SKM 是一个简单而强大的 SSH 密钥管理工具。它帮助您轻松管理多个 SSH 密钥！

[English README](README.md)

![](https://github.com/TimothyYe/skm/blob/master/assets/snapshots/demo.gif?raw=true)

## 功能

* 创建、列出、删除您的 SSH 密钥
* 通过别名管理所有 SSH 密钥
* 选择并设置默认 SSH 密钥
* 通过别名显示公钥
* 将任意 SSH 密钥复制到远程主机（支持 `--key`、`--pick`、`--dry-run`）
* 重命名 SSH 密钥别名
* 备份和恢复所有 SSH 密钥
* 从磁盘上的任意位置导入现有密钥对
* 将单个密钥导出为（可选加密的）打包文件
* 使用 `fingerprint` 与 `info` 查看密钥详情
* 添加 / 更换 / 移除密钥的口令
* 使用 `skm doctor` 诊断您的环境
* 多命令共享的提示界面（带模糊搜索）选择 SSH 密钥
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
     cache            Add your SSH to SSH agent cache via alias name.
     help, h          Shows a list of commands or help for one command.

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

默认输出为一个列对齐的表格，包含密钥类型、位数、指纹、是否已加载到 `ssh-agent`、修改日期以及注释。当前的默认密钥会被高亮显示。

```bash
% skm ls

✔ Found 3 SSH key(s)!

  ALIAS    TYPE     BITS  FINGERPRINT                                         AGENT  CREATED     COMMENT
* default  ed25519  256   SHA256:pFsC7J7L9L08f3w8uP6ozRaGW5Dg8CdEkP8iVj7++pw  yes    2026-05-16  work@laptop
  dev      ed25519  256   SHA256:7dFJEj7WGAL8rn9AqLNYdoTQrqgv00kdnqJlufvxgg4  -      2026-05-12  dev
  prod     rsa      4096  SHA256:DEyhI38hQ5WYABdx9SrJuhqrIyLvfRcZTtzXARuyn0k  -      2026-05-01  prod
```

其他常用选项：

```bash
skm ls -q              # 简洁模式 —— 仅输出别名，默认密钥前用 -> 标记
skm ls --json          # 适合脚本使用的机器可读输出
skm ls -t ed25519      # 按密钥类型过滤
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

默认情况下会推送当前激活的密钥：

```bash
% skm cp timothy@example.com

/usr/bin/ssh-copy-id: INFO: Source of key(s) to be installed: "/Users/timothy/.skm/default/id_rsa.pub"
...
✔  SSH key copied to remote host
```

常用选项：

```bash
skm cp --key work timothy@example.com         # 推送指定的密钥而非默认密钥
skm cp --pick timothy@example.com             # 交互式选择密钥
skm cp -p 2222 timothy@example.com            # 非默认 SSH 端口
skm cp timothy@[2001:db8::1]:2222             # 同样支持 IPv6 主机
skm cp --dry-run timothy@example.com          # 预览将要执行的 ssh-copy-id 命令
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

### 导入已有密钥

将磁盘上任意位置的密钥对导入到 SKM 存储中。SKM 会通过 `.pub` 后缀推断密钥对的另一半，并从公钥头部识别密钥类型。

```bash
% skm import --alias work ~/old-laptop/.ssh/id_ed25519
✔ Imported ed25519 key as [work]
```

`skm import` 还支持直接导入 `skm export` 生成的打包文件（`.tar.gz`、`.tgz` 或 `.tar.gz.enc`）。对于加密的打包文件，`openssl` 会提示输入口令。

```bash
% skm import ~/skm-work-20260516193629.tar.gz.enc
enter aes-256-cbc decryption password:
✔ Imported bundle as [work]
```

常用选项：

```bash
skm import --alias newname bundle.tar.gz      # 导入时重命名别名
skm import --move ~/old/id_ed25519            # 导入成功后删除源文件
```

> **注意：** 请将 `--alias` / `--move` 等参数放在路径之前 —— 放在路径之后的标志不会被解析，遇到这种情况 SKM 会报错并给出提示。

### 导出单个密钥

将一个别名打包成可移植的归档文件，便于在另一台机器上导入。归档中会包含该别名下的私钥、公钥以及任何 `hook` 文件。

```bash
% skm export work
✔ Exported [work] to /Users/timothy/skm-work-20260516193629.tar.gz
```

加上 `--encrypt` 可以使用 `openssl enc -aes-256-cbc -pbkdf2` 对打包文件进行加密，命令会提示输入口令：

```bash
% skm export --encrypt work
enter aes-256-cbc encryption password:
Verifying - enter aes-256-cbc encryption password:
✔ Exported [work] to /Users/timothy/skm-work-20260516193629.tar.gz.enc
```

也可以指定输出路径：

```bash
skm export -o /tmp/work.tar.gz work
```

### 查看密钥信息

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

未提供别名时，这两个命令都会默认作用于当前激活的密钥。

### 添加 / 更换 / 移除口令

`skm passphrase` 封装了 `ssh-keygen -p`。在新口令处输入空内容即可移除现有口令。

```bash
% skm passphrase work
Updating passphrase for [work] (/Users/timothy/.skm/work/id_ed25519)
Enter new passphrase (empty for no passphrase):
Enter same passphrase again:
Your identification has been saved with the new passphrase.
✔ Passphrase updated for [work]
```

### 运行环境诊断

`skm doctor` 会自动检查最常见的环境问题 —— SSH 工具是否缺失、`ssh-agent` 是否可达、存储路径是否可写、默认密钥是否解析成功、RSA 位数是否低于 3072、公私钥文件权限是否过松，以及 hook 脚本是否具有可执行位。

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

`skm doctor --json` 会以 JSON 数组的形式输出相同的检查结果，便于脚本处理。失败项会导致非零退出码；警告不会。

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

## Star 历史

[![Star History Chart](https://api.star-history.com/svg?repos=timothyye/skm&type=date&legend=top-left)](https://www.star-history.com/#timothyye/skm&type=date&legend=top-left)

## 许可证

[MIT 许可证](https://github.com/TimothyYe/skm/blob/master/LICENSE)  
