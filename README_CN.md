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

* 创建、列出、删除您的 SSH 密钥（支持 ed25519、rsa、ed25519-sk、ecdsa-sk）
* 通过别名管理所有 SSH 密钥
* 选择并设置默认 SSH 密钥
* 通过别名显示公钥
* 将任意 SSH 密钥复制到远程主机（支持 `--key`、`--pick`、`--dry-run`）
* 重命名 SSH 密钥别名
* 备份和恢复所有 SSH 密钥（支持 `--encrypt` 加密备份)
* 使用 `skm trash list|restore|empty` 软删除并恢复误删的密钥
* 从磁盘上的任意位置导入现有密钥对
* 将单个密钥导出为（可选加密的）打包文件
* 使用 `fingerprint` 与 `info` 查看密钥详情
* 添加 / 更换 / 移除密钥的口令
* 使用 `skm doctor` 诊断您的环境
* 使用 `skm audit` 审计已存储密钥的强度、口令保护与年龄
* 多命令共享的提示界面（带模糊搜索）选择 SSH 密钥
* 自定义 SSH 密钥存储路径
* 可插拔的钩子机制：支持 `post-use`、`post-create`、`pre-delete`、`post-copy` 事件（按密钥与全局两级）

## 安装

#### Homebrew

从 **v0.8.9** 版本开始，`skm` 已被正式收录到 [homebrew-core](https://github.com/Homebrew/homebrew-core) 官方仓库，您可以直接在 macOS 和 Linux 上安装：

```bash
brew install skm
```

> 如果您之前通过旧的 tap 安装过 `skm`，请先移除以避免冲突：
>
> ```bash
> brew uninstall skm
> brew untap timothyye/tap
> brew install skm
> ```

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
     init, i          Initialize SSH keys store for the first time use
     create, c        Create a new SSH key (defaults to ed25519; rsa, ed25519-sk, ecdsa-sk also supported)
     ls, l            List all the available SSH keys (alias + comment by default; use -l for full details)
     use, u           Set specific SSH key as default by its alias name
     delete, d        Delete specific SSH key by alias name (moves to trash; restore with `skm trash restore`)
     rename, rn       Rename SSH key alias name to a new one
     copy, cp         Copy SSH public key to a remote host
     display, dp      Display the current SSH public key or specific SSH public key by alias name
     backup, b        Backup all SSH keys to an archive file (or a restic repository with --restic)
     restore, r       Restore SSH keys from an existing archive file (.enc bundles are auto-decrypted)
     import, im       Import an existing SSH key pair from a path into the store
     export, ex       Export a single key as a tar.gz bundle (optionally encrypted)
     fingerprint, fp  Print the SHA256 fingerprint of an SSH key (default: active key)
     info, in         Show detailed information about an SSH key (default: active key)
     passphrase, pp   Add, rotate, or remove the passphrase on an SSH key
     doctor, dr       Run diagnostics against the SKM environment, agent, and stored keys
     audit, au        Audit stored keys for weak strength, missing passphrases, and age
     trash            Manage soft-deleted SSH keys
     hook             Inspect hooks wired to SKM events
     cache            Add your SSH to SSH agent cache via alias name
     help, h          Shows a list of commands or help for one command

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

支持的密钥类型：`ed25519`（默认）、`rsa`、`ed25519-sk`、`ecdsa-sk`。`-sk` 变体是 FIDO2 硬件支持的密钥，创建时需要 `ssh-keygen` 8.2+ 以及插入的硬件密钥（YubiKey、Solo 等）。

```bash
% skm create prod -C "abc@abc.com"
Generating public/private ed25519 key pair.
Enter passphrase (empty for no passphrase):
Enter same passphrase again:
Your identification has been saved in /Users/timothy/.skm/prod/id_ed25519.
Your public key has been saved in /Users/timothy/.skm/prod/id_ed25519.pub.
...
✔ SSH key [prod] created!
```

其他示例：

```bash
skm create old -t rsa -b 4096           # RSA 密钥（最低 3072 位）
skm create yubi -t ed25519-sk           # 硬件支持的密钥；会提示 PIN 与触摸
```

低于 3072 位的 RSA 密钥会被直接拒绝 —— `skm audit` 也会把这些标为弱密钥，所以 `create` 与 `audit` 在"什么算安全"上保持一致。若你确实需要更小的密钥用于测试，请直接使用 `ssh-keygen`。

### 列出 SSH 密钥

默认情况下 `ls` 以三列输出每个别名 —— 别名、密钥类型、注释；当前激活的默认密钥以 `->` 标记。使用 `-l` / `--long` 查看完整表格，包含位数、指纹、agent 状态以及修改日期。

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
✔ SSH key [prod] moved to trash (restore with: skm trash restore prod-20260521150412)
```

默认情况下，删除会把别名移入存储中的回收站，方便误删后恢复。一次传入多个别名即可批量删除；不存在的别名会被报告并跳过。`-y` / `--yes` 跳过确认提示，`--purge` 直接硬删除（不进入回收站）：

```bash
% skm delete -y staging legacy old-laptop
% skm delete --purge --yes ancient
```

### 恢复已删除的 SSH 密钥

```bash
% skm trash list
NAME                       ALIAS    DELETED
prod-20260521150412        prod     2026-05-21 15:04:12

% skm trash restore prod-20260521150412
✔ Restored [prod-20260521150412] as alias [prod]
```

参数既可以传入回收站条目名（`prod-20260521150412`），也可以只传别名（`prod`）—— 当只有一个匹配条目时直接命中。如果原始别名已被占用，请使用 `--as <new-alias>`。使用 `skm trash empty` 永久清空回收站（会提示确认，`-y` 跳过）。

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

将所有 SSH 密钥备份为 `$HOME` 中的一个 tar 包。

```bash
% skm backup

a .
a ./default/id_rsa
a ./default/id_rsa.pub
…
✔  All SSH keys backup to: /Users/timothy/skm-20260521221544.tar.gz
⚠ This bundle contains UNENCRYPTED private keys. If it leaves this
  machine, anyone with the file can use your keys. Re-run with
  --encrypt to produce an encrypted archive.
```

默认 tar 包里的私钥是明文。加上 `--encrypt` 可生成加密版本（AES-256-CBC，envelope 与 `skm export --encrypt` 相同）：

```bash
% skm backup --encrypt
enter AES-256-CBC encryption password:
Verifying - enter AES-256-CBC encryption password:
✔  All SSH keys backup to: /Users/timothy/skm-20260521221544.tar.gz.enc
  Decrypt with: openssl enc -d -aes-256-cbc -pbkdf2 -in /Users/timothy/skm-20260521221544.tar.gz.enc -out skm-20260521221544.tar.gz
```

脚本场景可用 `--password-file <path>` 从文件读取口令而不弹出提示。

#### 基于 restic 的备份（本地或云端）

如果你装了 [restic](https://restic.net/)，SKM 可以借助它生成加密、去重、快照式的备份，能直接写入本地磁盘、S3、Cloudflare R2、Backblaze B2、SFTP，以及任意 restic 支持的后端。第一次使用前请先运行交互式初始化：

```bash
% skm backup --restic --init
Configuring restic backup for SKM.
Examples:
  Local:  /Users/me/.skm-backups
  S3:     s3:s3.amazonaws.com/my-bucket/skm
  R2:     s3:https://<account>.r2.cloudflarestorage.com/my-bucket/skm
  SFTP:   sftp:user@host:/data/skm
  B2:     b2:my-bucket/skm

For S3, R2, and B2, set the relevant credential env vars (e.g.
AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY) before running backup.

✔ Restic repository initialized at s3:https://abc.r2.cloudflarestorage.com/skm
⚠ IMPORTANT: store the restic password somewhere OTHER than this machine.
  If this laptop is lost and the password only lives in ~/.skm-backups.passwd,
  the backup will be unrecoverable.
```

之后只需要：

```bash
% skm backup --restic
✔  Backup to s3:https://abc.r2.cloudflarestorage.com/skm complete
```

restic 会在客户端完成所有分块和加密，**远端只会看到密文**。R2/S3 的凭据与 restic 加密口令是两组独立的秘密 —— 凭据丢了可以重发，restic 口令丢了备份就再也解不开了，所以请务必把它存到机器之外的地方。

### 恢复 SSH 密钥

```bash
% skm restore ~/skm-20260521221544.tar.gz
✔  All SSH keys restored to /Users/timothy/.skm
```

`.enc` 加密包会自动识别并解密：

```bash
% skm restore ~/skm-20260521221544.tar.gz.enc
enter AES-256-CBC decryption password:
✔  All SSH keys restored to /Users/timothy/.skm
```

对于 restic 备份，指定要恢复的快照：

```bash
% skm restore --restic --restic-snapshot $SNAPSHOT
✔  Backup restored to /Users/$USER/.skm
```

省略 `--restic-snapshot` 时，SKM 会列出所有可用的快照供你选择。

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

### 审计已存储的密钥

`doctor` 关注环境，而 `skm audit`（别名 `au`）关注密钥本身。它会遍历存储中的每一个密钥并报告：

* RSA 位数低于 `--rsa-min`（默认 3072）—— **失败**
* 私钥未设置口令保护 —— **警告**
* 密钥年龄超过 `--max-age`（默认 `1y`，支持 `Nd|Nw|Nm|Ny`）—— **警告**

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

可用参数：

* `--json` —— 以 JSON 数组形式输出结果，便于脚本处理
* `--strict` —— 将所有警告提升为失败（适合在 CI 中使用）
* `--max-age <duration>` —— 自定义年龄阈值（如 `30d`、`6m`）
* `--rsa-min <bits>` —— 自定义可接受的最小 RSA 位数

若存在任何失败级别的结果，`audit` 会以非零状态码退出；仅有警告时返回 0，除非使用 `--strict`。

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

SKM 会在密钥生命周期事件前后触发钩子脚本。钩子是普通的可执行文件（二进制或脚本），只需放到约定位置，SKM 会自动调用。

#### 事件列表

| 事件          | 触发时机                                | 失败行为                            |
| ------------- | --------------------------------------- | ----------------------------------- |
| `post-use`    | 切换默认 SSH 密钥之后                   | 尽力执行（仅打印警告）              |
| `post-create` | `skm create` 成功之后                   | 尽力执行（仅打印警告）              |
| `pre-delete`  | 用户确认删除后、实际移除之前            | **退出码非零将中止删除**            |
| `post-copy`   | `skm copy`（ssh-copy-id）成功之后       | 尽力执行（仅打印警告）              |

#### 钩子脚本的存放位置

钩子分为两个级别，会按顺序都触发（先全局，再按密钥）：

```bash
~/.skm/hooks/<event>              # 全局：对任何别名都触发
~/.skm/<alias>/hooks/<event>      # 按密钥：仅对该别名触发
```

示例：

```bash
~/.skm/hooks/pre-delete           # 所有删除操作的全局守卫
~/.skm/work/hooks/post-use        # 切换到 "work" 时才触发
~/.skm/prod/hooks/post-copy       # 复制 "prod" 密钥后才触发
```

旧版的 `~/.skm/<alias>/hook` 文件依然受支持，会被当作 `post-use` 钩子，方便向后兼容。

请记得为脚本加上可执行权限：

```bash
chmod +x ~/.skm/work/hooks/post-use
```

#### 传给钩子的环境变量

每个钩子在执行时都能读取以下环境变量：

| 变量               | 含义                                           |
| ------------------ | ---------------------------------------------- |
| `SKM_EVENT`        | 事件名称（`post-use`、`pre-delete` 等）        |
| `SKM_ALIAS`        | 密钥别名（同时作为 `$1` 传入）                 |
| `SKM_STORE_PATH`   | SKM 存储路径                                   |
| `SKM_SSH_PATH`     | 用户的 `~/.ssh` 路径                           |
| `SKM_KEY_TYPE`     | `rsa`、`ed25519` 等                            |
| `SKM_PRIVATE_KEY`  | 私钥的绝对路径                                 |
| `SKM_PUBLIC_KEY`   | 公钥的绝对路径                                 |

事件相关的额外变量：

| 事件        | 额外变量                                |
| ----------- | --------------------------------------- |
| `post-copy` | `SKM_REMOTE_HOST`、`SKM_REMOTE_PORT`    |

#### 示例：`post-use` 时切换 git 身份

```bash
#!/bin/bash
# ~/.skm/work/hooks/post-use
git config --global user.name  "你的工作姓名"
git config --global user.email "you@work.example.com"
```

#### 示例：`post-copy` 时记录每次密钥分发

```bash
#!/bin/sh
# ~/.skm/hooks/post-copy
echo "$(date -Iseconds) $SKM_ALIAS -> $SKM_REMOTE_HOST:${SKM_REMOTE_PORT:-22}" \
  >> ~/.skm/deployments.log
```

#### 示例：通过 `pre-delete` 防止误删生产密钥

```bash
#!/bin/sh
# ~/.skm/prod/hooks/pre-delete —— 非零退出会中止删除
echo "拒绝删除生产密钥，请先移除此钩子。" >&2
exit 1
```

#### 查看已配置的钩子

```bash
skm hook ls              # 仅列出全局钩子
skm hook ls <alias>      # 列出全局 + 该别名的按密钥钩子
skm hook ls --all        # 列出全局 + 所有别名的按密钥钩子
```

## Star 历史

[![Star History Chart](https://api.star-history.com/svg?repos=timothyye/skm&type=date&legend=top-left)](https://www.star-history.com/#timothyye/skm&type=date&legend=top-left)

## 许可证

[MIT 许可证](https://github.com/TimothyYe/skm/blob/master/LICENSE)  
