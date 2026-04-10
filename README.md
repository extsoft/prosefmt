# prosefmt

![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)

`prosefmt` is the simplest text formatter for making your files look correct. No complex rules,
no massive configuration files — just clean text.

## Contents

- [Overview](#overview)
- [Getting Started](#getting-started)
  - [Installation](#installation)
  - [Usage](#usage)
- [CLI Reference](#cli-reference)
- [Development](#development)
- [License](#license)

## Overview

Over the years, I’ve gotten used to formatting text and code files in a specific way, such as removing
trailing spaces and ensuring there’s a newline at the end of each file. In many projects, a code formatter
handles these tasks. However, some projects simply don’t use one, and introducing a formatter can be
a significant challenge. Even when a formatter is configured, some files may still be ignored because most
formatters are designed for specific languages and file types only.

`prosefmt` is a simple formatter for any text file.
If a project doesn’t have a formatter, it can be the first one you introduce.
If a project already uses other formatters, `prosefmt` can be a useful addition.

`prosefmt` is designed to process any text files while automatically ignoring binary files (those containing
null bytes or control characters). It does not use configuration files; all settings are specified via
the command line arguments with sensible defaults. For security reasons, files are only overwritten when
the `write` command is explicitly used.

The tool supports the following rules:

- `PF1`: a file must end with exactly one newline.
- `PF2`: no trailing space(s) or tab(s) at the end of a line.
- `PF3`: detect and preserve existing line endings (LF or CRLF); default behavior.
- `PF4`: enforce Linux-style LF (`\n`) line endings.
- `PF5`: enforce Windows-style CRLF (`\r\n`) line endings.
- `PF6`: when [`--replace-tabs-with-spaces`](#--replace-tabs-with-spaces) is set, tabs must be replaced with that many spaces.
- `PF7`: when [`--replace-spaces-with-tabs`](#--replace-spaces-with-tabs) is set, each run of that many spaces must be replaced with a tab.

## Getting Started

### Installation

#### Install Script

Download and install

- the latest release: `curl -fsSL https://prosefmt.extsoft.pro/install.sh | sh`
- to a custom directory: `curl -sSL https://prosefmt.extsoft.pro/install.sh | sh -s -- -d ~/.local/bin`
- a specific version: `curl -sSL https://prosefmt.extsoft.pro/install.sh | sh -s -- -v v1.0.0`

#### `mise`

```sh
mise use github:extsoft/prosefmt
```

#### Go Install

```sh
go install github.com/extsoft/prosefmt@latest
```

### Usage

For safety and to prevent unwanted file changes, run the `check` command first.

```sh
prosefmt check some/path some.file
```

The output shows which files will be updated and why. Once you’re ready to apply the changes, run the `write` command.

```sh
prosefmt write some/path some.file
```

## CLI Reference

**Usage**: `prosefmt [command] [file...]`

By default, `prosefmt` runs the [check](#check-command) command on the specified files
and directories — at least one file must be provided. Directories are scanned recursively.

**Commands**

- [check](#check-command) (default)
- [write](#write-command)
- [version](#version-command)
- [completion](#completion-command)

### `check` command

**Usage**: `prosefmt check [flags] files...`

The `check` command scans the specified files for formatting issues. Binary files are ignored.
If a directory is provided, it is scanned recursively to find files.
The command exits with code `1` if at least one issue is detected; otherwise, it exits with code `0`.
The `check` command runs by default when no other command is specified.

**Flags**

- [Configuration Flags](#configuration-flags)
  - [--line-endings](#--line-endings)
  - [--replace-tabs-with-spaces](#--replace-tabs-with-spaces)
  - [--replace-spaces-with-tabs](#--replace-spaces-with-tabs)
- [Output Flags](#output-flags)
  - [--silent](#--silent)
  - [--compact](#--compact)
  - [--verbose](#--verbose)

### `write` command

**Usage**: `prosefmt write [flags] files...`

The `write` command fixes formatting issues in the specified files. Binary files are ignored.
If a directory is provided, it is scanned recursively to find files.
The exit code is always `0`.

**Flags**

- [Configuration Flags](#configuration-flags)
  - [--line-endings](#--line-endings)
  - [--replace-tabs-with-spaces](#--replace-tabs-with-spaces)
  - [--replace-spaces-with-tabs](#--replace-spaces-with-tabs)
- [Output Flags](#output-flags)
  - [--silent](#--silent)
  - [--compact](#--compact)
  - [--verbose](#--verbose)

### `version` command

**Usage**: `prosefmt version`

The `version` command prints the tool version.

### `completion` command

**Usage**: `prosefmt completion bash|zsh|fish|powershell`

The `completion` command generates shell completion scripts.

### Configuration Flags

#### `--line-endings`

> This flag is available for the [check](#check-command) and [write](#write-command) commands only.

The `--line-endings` flag configures how line endings are handled.
Use `auto` (default) to preserve existing line endings. `linux` enforces LF (\n). `windows` enforces CRLF (\r\n).

#### `--replace-tabs-with-spaces`

> This flag is available for the [check](#check-command) and [write](#write-command) commands only.

The `--replace-tabs-with-spaces` flag takes a **positive integer** `N`. When set, each tab character (`\t`) in a file is treated as a formatting issue ([`PF6`](#overview)) and is replaced with exactly `N` space characters on [`write`](#write-command). Omit the flag to leave tab characters unchanged (default).

Values that are not positive integers (for example `0`, negative numbers, or non-numeric values) are rejected.

[`--replace-spaces-with-tabs`](#--replace-spaces-with-tabs) and `--replace-tabs-with-spaces` are **mutually exclusive**; only one may be passed.

#### `--replace-spaces-with-tabs`

> This flag is available for the [check](#check-command) and [write](#write-command) commands only.

The `--replace-spaces-with-tabs` flag takes a **positive integer** `N`. When set, each run of **N** consecutive space characters is treated as a formatting issue ([`PF7`](#overview)) and is replaced with a tab character on [`write`](#write-command). Replacement repeats until no run of `N` spaces remains (for example, eight spaces with `N` = 4 become two tabs). Omit the flag to leave space characters unchanged (default).

Values that are not positive integers (for example `0`, negative numbers, or non-numeric values) are rejected.

[`--replace-tabs-with-spaces`](#--replace-tabs-with-spaces) and `--replace-spaces-with-tabs` are **mutually exclusive**; only one may be passed.

### Output Flags

The [check](#check-command) and [write](#write-command) commands print output to standard output.
The flags below determine the output format.

> These flags are available for the [check](#check-command) and [write](#write-command) commands only.

#### `--silent`

No output printed.

#### `--compact`

The compact output format prints issue details for the [check](#check-command) command or
updated file names for the [write](#write-command) command.

#### `--verbose`

The verbose output format provides detailed debug information, including processing steps,
a scanning summary, scanner acceptance and rejection decisions with reasons, rules applied per file,
write operations, and timing information.

## Development

Install [mise](https://mise.jdx.dev/) (dev tool version manager), then run `mise run init` in the repo so the project’s tools and env are activated. See [mise docs](https://mise.jdx.dev/) for install and usage.

This project uses [hk](https://hk.jdx.dev) for code checks and git hooks. Use `mise run check` or `mise run fix` to check or autofix.

`mise run build` builds the CLI binary.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
