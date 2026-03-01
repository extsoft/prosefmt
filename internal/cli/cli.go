package cli

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/extsoft/prosefmt/internal/fix"
	"github.com/extsoft/prosefmt/internal/log"
	"github.com/extsoft/prosefmt/internal/report"
	"github.com/extsoft/prosefmt/internal/rules"
	"github.com/extsoft/prosefmt/internal/scanner"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	Version        = "dev"
	checkHadIssues bool
)

const rootDescription = "The simplest text formatter to keep your files consistently formatted."
const helpTextWidth = 72

var rootCmd = &cobra.Command{
	Use:   "prosefmt [command]",
	Short: rootDescription,
	Long:  rootDescription,
	Args:  cobra.ArbitraryArgs,
	RunE:  rootRunE,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the tool version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(Version)
	},
}

var checkCmd = &cobra.Command{
	Use:   "check [flags] files...",
	Short: "Find formatting issues in the given file(s)",
	Long: "The `check` command scans the specified files for formatting issues. " +
		"Binary files are ignored. If a directory is provided, it is scanned recursively to find files. " +
		"The command exits with code `1` if at least one issue is detected; otherwise, it exits with code `0`. " +
		"The `check` command runs by default when no other command is specified.",
	Args: cobra.ArbitraryArgs,
	RunE: checkRunE,
}

var writeCmd = &cobra.Command{
	Use:   "write [flags] files...",
	Short: "Find formatting issues in the given file(s)",
	Long: "The `write` command fixes formatting issues in the specified files. " +
		"Binary files are ignored. If a directory is provided, it is scanned recursively to find files. " +
		"The exit code is always `0`.",
	Args: cobra.ArbitraryArgs,
	RunE: writeRunE,
}

func addOutputFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("silent", false, "No output printed")
	cmd.Flags().Bool("compact", false, "Show formatted or errored files (default)")
	cmd.Flags().Bool("verbose", false, "Print debug output (steps, scanner, rules, timing)")
}

func addLineEndingsFlag(cmd *cobra.Command) {
	cmd.Flags().String("line-endings", "auto", "Use `auto` (default) to preserve existing line endings. `linux` enforces LF (\\n). `windows` enforces CRLF (\\r\\n).")
}

func lineEndingsModeFromCmd(cmd *cobra.Command) (rules.LineEndingMode, error) {
	s, _ := cmd.Flags().GetString("line-endings")
	return rules.ParseLineEndingMode(s)
}

func outputLevelFromCmd(cmd *cobra.Command) log.Level {
	silent, _ := cmd.Flags().GetBool("silent")
	compact, _ := cmd.Flags().GetBool("compact")
	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		return log.Verbose
	}
	if compact {
		return log.Normal
	}
	if silent {
		return log.Silent
	}
	return log.Normal
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(writeCmd)
	addOutputFlags(checkCmd)
	addOutputFlags(writeCmd)
	addLineEndingsFlag(checkCmd)
	addLineEndingsFlag(writeCmd)
	rootCmd.SetHelpFunc(rootHelpFunc)
	checkCmd.SetHelpFunc(commandHelpFunc)
	writeCmd.SetHelpFunc(commandHelpFunc)
}

var outputFlagOrder = []string{"silent", "compact", "verbose"}

func rootHelpFunc(cmd *cobra.Command, args []string) {
	width, _ := cmd.Flags().GetInt("help-width")
	out := cmd.OutOrStderr()
	fmt.Fprintf(out, "Usage: %s\n", cmd.UseLine())
	if cmd.Short != "" {
		fmt.Fprintf(out, "\n%s\n", wrapWords(cmd.Short, width))
	}
	if len(cmd.Commands()) > 0 {
		fmt.Fprintln(out, "\nCommands:")
		maxLen := 0
		for _, c := range cmd.Commands() {
			if c.IsAvailableCommand() && len(c.Name()) > maxLen {
				maxLen = len(c.Name())
			}
		}
		for _, c := range cmd.Commands() {
			if c.IsAvailableCommand() {
				usage := wrapWords(c.Short, width-maxLen-4)
				lines := strings.Split(usage, "\n")
				fmt.Fprintf(out, "  %-*s  %s\n", maxLen, c.Name(), lines[0])
				for _, line := range lines[1:] {
					fmt.Fprintf(out, "  %-*s  %s\n", maxLen, "", line)
				}
			}
		}
	}
	fmt.Fprintln(out, "\nIf no command is specified, `check` runs by default.")
	if Version != "" {
		fmt.Fprintf(out, "\nVersion: %s\n", Version)
	}
}

func wrapWords(s string, width int) string {
	if width <= 0 {
		return s
	}
	var b strings.Builder
	for _, para := range strings.Split(s, "\n") {
		para = strings.TrimSpace(para)
		if para == "" {
			b.WriteString("\n")
			continue
		}
		for {
			if len(para) <= width {
				b.WriteString(para)
				b.WriteString("\n")
				break
			}
			i := strings.LastIndex(para[:width+1], " ")
			if i <= 0 {
				i = width
			}
			b.WriteString(para[:i])
			b.WriteString("\n")
			para = strings.TrimSpace(para[i:])
		}
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func commandHelpFunc(cmd *cobra.Command, args []string) {
	out := cmd.OutOrStderr()
	fmt.Fprintf(out, "Usage: %s\n", cmd.UseLine())
	if cmd.Long != "" {
		fmt.Fprintf(out, "\n%s\n", wrapWords(cmd.Long, helpTextWidth))
	}
	if f := cmd.Flags().Lookup("line-endings"); f != nil {
		fmt.Fprintln(out, "\nConfiguration:")
		printFlagUsage(out, f, helpTextWidth)
	}
	fmt.Fprintln(out, "\nOutput:")
	for _, name := range outputFlagOrder {
		if f := cmd.Flags().Lookup(name); f != nil {
			printFlagUsage(out, f, helpTextWidth)
		}
	}
	if f := cmd.Flags().Lookup("help-width"); f != nil {
		fmt.Fprintln(out, "\nHelp:")
		printFlagUsage(out, f, helpTextWidth)
	}
	if Version != "" {
		fmt.Fprintf(out, "\nVersion: %s\n", Version)
	}
}

func printFlagUsage(out io.Writer, f *pflag.Flag, width int) {
	var prefix string
	if f.Shorthand != "" && f.Name != f.Shorthand {
		prefix = fmt.Sprintf("  -%s, --%s", f.Shorthand, f.Name)
	} else {
		prefix = fmt.Sprintf("      --%s", f.Name)
	}

	usageWidth := width - len(prefix) - 2
	if usageWidth < 10 {
		usageWidth = 10
	}
	usage := wrapWords(f.Usage, usageWidth)
	lines := strings.Split(usage, "\n")
	fmt.Fprintf(out, "%s  %s\n", prefix, lines[0])
	for _, line := range lines[1:] {
		fmt.Fprintf(out, "%*s  %s\n", len(prefix), "", line)
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if checkHadIssues {
		os.Exit(1)
	}
}

func rootRunE(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		rootHelpFunc(cmd, nil)
		return nil
	}
	log.SetLevel(log.Normal)
	hadIssues, err := Run(true, false, args, rules.LineEndAuto)
	if err != nil {
		return err
	}
	checkHadIssues = hadIssues
	return nil
}

func checkRunE(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		commandHelpFunc(cmd, nil)
		return nil
	}
	mode, err := lineEndingsModeFromCmd(cmd)
	if err != nil {
		return err
	}
	log.SetLevel(outputLevelFromCmd(cmd))
	hadIssues, err := Run(true, false, args, mode)
	if err != nil {
		return err
	}
	checkHadIssues = hadIssues
	return nil
}

func writeRunE(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		commandHelpFunc(cmd, nil)
		return nil
	}
	mode, err := lineEndingsModeFromCmd(cmd)
	if err != nil {
		return err
	}
	log.SetLevel(outputLevelFromCmd(cmd))
	_, err = Run(false, true, args, mode)
	return err
}

func Run(check, doWrite bool, files []string, mode rules.LineEndingMode) (hadIssues bool, err error) {
	start := time.Now()
	lvl := log.GetLevel()
	if lvl >= log.Verbose {
		log.Logf(log.Verbose, "Configuration: check=%v files=%v\n", check, files)
	}
	scanned, skipped, err := scanner.Scan(files)
	if err != nil {
		return false, err
	}
	elapsedScan := time.Since(start)
	if lvl >= log.Verbose {
		if len(scanned) == 0 {
			log.Logf(log.Verbose, "No text files found. Scanned 0 text file(s), skipped %d argument(s).\n", len(skipped))
		} else {
			log.Logf(log.Verbose, "Scanned %d text file(s), skipped %d argument(s).\n", len(scanned), len(skipped))
		}
		for _, p := range sortedKeys(skipped) {
			log.Logf(log.Verbose, "scanner: rejected %s (reason: %s)\n", p, skipped[p])
		}
		for _, p := range scanned {
			log.Logf(log.Verbose, "scanner: accepted %s\n", p)
		}
	}
	if len(scanned) == 0 {
		if lvl >= log.Normal {
			fmt.Fprintln(os.Stdout, "No text files found.")
			if check {
				report.Write(os.Stdout, report.FormatCompact, nil, 0, nil)
			}
		}
		return false, nil
	}
	var allIssues []rules.Issue
	fileIssues := make(map[string][]rules.Issue)
	for _, path := range scanned {
		if lvl >= log.Verbose {
			if check {
				log.Logf(log.Verbose, "Checking %s\n", path)
			} else {
				log.Logf(log.Verbose, "Writing %s\n", path)
			}
		}
		issues, err := rules.CheckFile(path, mode)
		if err != nil {
			return false, err
		}
		if len(issues) > 0 {
			fileIssues[path] = issues
			allIssues = append(allIssues, issues...)
			if lvl >= log.Verbose {
				ruleIDs := make(map[string]bool)
				for _, i := range issues {
					ruleIDs[i.RuleID] = true
				}
				var ids []string
				for id := range ruleIDs {
					ids = append(ids, id)
				}
				sort.Strings(ids)
				log.Logf(log.Verbose, "rules: %s -> %d issue(s): %s\n", path, len(issues), strings.Join(ids, ", "))
			}
		}
	}
	if check {
		if lvl >= log.Normal {
			if err := report.Write(os.Stdout, report.FormatCompact, allIssues, len(scanned), scanned); err != nil {
				return false, err
			}
		}
		elapsed := time.Since(start)
		if lvl >= log.Verbose {
			log.Logf(log.Verbose, "Completed in %s\n", elapsed.Round(time.Millisecond))
		}
		_ = elapsedScan
		return len(allIssues) > 0, nil
	}
	for path := range fileIssues {
		if err := fix.Apply(path, mode); err != nil {
			return false, err
		}
		if lvl >= log.Verbose {
			log.Logf(log.Verbose, "write: applied to %s\n", path)
		}
	}
	if lvl >= log.Normal && len(fileIssues) > 0 {
		written := make([]string, 0, len(fileIssues))
		for p := range fileIssues {
			written = append(written, p)
		}
		sort.Strings(written)
		fmt.Fprintf(os.Stdout, "Wrote %d file(s):\n", len(written))
		for _, p := range written {
			fmt.Fprintln(os.Stdout, p)
		}
	}
	elapsed := time.Since(start)
	if lvl >= log.Verbose {
		log.Logf(log.Verbose, "Completed in %s\n", elapsed.Round(time.Millisecond))
	}
	return false, nil
}

func sortedKeys(m map[string]string) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
