# prosefmt

![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)

`prosefmt` is the simplest text formatter for making your files look correct. No complex rules, no massive configuration files — just clean text.

## CLI reference

**Synopsis**

```bash
prosefmt [command] [flags] [path...]
```

Pass at least one file or directory; directories are scanned recursively. By default the tool runs in check mode (report only). Use `--write` to apply fixes in place.

**Commands**

- [version](#version)
- [completion](#completion)

**Options**

- [--check](#--check)
- [--write](#--write)

**Output**

- [--silent](#--silent)
- [--compact](#--compact)
- [--verbose](#--verbose)

### `version`

Print the version number. Run: `prosefmt version`.

### `completion`

Generate a shell completion script. Usage: `prosefmt completion <shell>` with one of `bash`, `zsh`, `fish`, or `powershell`. See [Shell completion](#shell-completion) below for install steps.

### Options

#### `--check`

Check only: scan paths and report issues to stdout. Exit code is 1 if any issue is found, 0 otherwise. This is the default when neither `--check` nor `--write` is set. Exactly one of `--check` or `--write` is allowed.

#### `--write`

Write fixes in place. Files with issues are modified on disk. Prints how many files were written and lists each path; exit code is 0. Exactly one of `--check` or `--write` is allowed.

### Output

Check mode prints a compact report: one line per issue as `file:line:col: rule: message`, grouped by file then rule; then a summary line `N file(s) scanned, M issue(s).`

By default output is **compact**: report (or formatted/errored file summary), "No text files found." when applicable, "Wrote N file(s):" plus one path per line in write mode. Use `--silent`, `--compact`, or `--verbose` to set the level explicitly.

#### `--silent`

No standard output printed. Exit code is still 1 when issues are found in check mode.

#### `--compact`

Show formatted or errored files: report in check mode, "No text files found." when applicable, "Wrote N file(s):" plus one path per line in write mode. This is the default when no output level flag is set. If multiple output flags are set, the noisiest wins (verbose > compact > silent).

#### `--verbose`

Print debug output: steps, scanning summary, scanner accepted/rejected with reasons, rules per file, write steps, and timing on stderr.

## Implementatio Notes

### Rules

| ID | Description |
|----|-------------|
| **TL001** | File must end with exactly one newline (LF or CRLF). |
| **TL010** | No trailing spaces or tabs at the end of a line. |

Both LF and CRLF line endings are supported; the tool preserves the detected style when writing.

### Text vs binary

Files are included only if they are valid UTF-8 and contain no null bytes. Binary and invalid-encoding files are skipped. When no text files are found, the summary includes "No text files found." (and "0 file(s) scanned, 0 issue(s).").

## Development

Install [mise](https://mise.jdx.dev/) (dev tool version manager), then run `mise run init` in the repo so the project’s tools and env are activated. See [mise docs](https://mise.jdx.dev/) for install and usage.

This project uses [hk](https://hk.jdx.dev) for code checks and git hooks. Use `mise run check` or `mise run fix` to check or autofix.

`mise run build` builds the CLI binary.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
