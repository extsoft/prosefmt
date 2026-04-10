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

// Option help layout matches common CLI style: flag on its own line, description
// indented on the following lines, blank line after each option.
const (
	optionLineIndent = "  "
	optionDescIndent = "      "
)

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
	Args:          cobra.ArbitraryArgs,
	RunE:          checkRunE,
	SilenceUsage:  true,
	SilenceErrors: true,
}

var writeCmd = &cobra.Command{
	Use:   "write [flags] files...",
	Short: "Find formatting issues in the given file(s)",
	Long: "The `write` command fixes formatting issues in the specified files. " +
		"Binary files are ignored. If a directory is provided, it is scanned recursively to find files. " +
		"The exit code is always `0`.",
	Args:          cobra.ArbitraryArgs,
	RunE:          writeRunE,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func addOutputFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("silent", false, "No output printed")
	cmd.Flags().Bool("compact", false, "Show formatted or errored files (default)")
	cmd.Flags().Bool("verbose", false, "Print debug output (steps, scanner, rules, timing)")
}

func addLineEndingsFlag(cmd *cobra.Command) {
	cmd.Flags().String("line-endings", "auto", "Use `auto` (default) to preserve existing line endings. `linux` enforces LF (\\n). `windows` enforces CRLF (\\r\\n).")
}

func addTabsToSpacesFlag(cmd *cobra.Command) {
	cmd.Flags().Int("replace-tabs-with-spaces", 0, "Replace each tab character with N space characters. Omit to leave tabs unchanged.")
}

func addSpacesToTabFlag(cmd *cobra.Command) {
	cmd.Flags().Int("replace-spaces-with-tabs", 0, "Replace each run of N space characters with a tab. Omit to leave spaces unchanged.")
}

func lineEndingsModeFromCmd(cmd *cobra.Command) (rules.LineEndingMode, error) {
	s, _ := cmd.Flags().GetString("line-endings")
	mode, err := rules.ParseLineEndingMode(s)
	if err != nil {
		return rules.LineEndAuto, optionFlagError("line-endings", err.Error())
	}
	return mode, nil
}

func tabWidthFromCmd(cmd *cobra.Command) (int, error) {
	if !cmd.Flags().Changed("replace-tabs-with-spaces") {
		return 0, nil
	}
	n, err := cmd.Flags().GetInt("replace-tabs-with-spaces")
	if err != nil {
		return 0, err
	}
	if n < 1 {
		return 0, optionFlagError("replace-tabs-with-spaces", "must be a positive integer")
	}
	return n, nil
}

func spacesToTabFromCmd(cmd *cobra.Command) (int, error) {
	if !cmd.Flags().Changed("replace-spaces-with-tabs") {
		return 0, nil
	}
	n, err := cmd.Flags().GetInt("replace-spaces-with-tabs")
	if err != nil {
		return 0, err
	}
	if n < 1 {
		return 0, optionFlagError("replace-spaces-with-tabs", "must be a positive integer")
	}
	return n, nil
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
	addTabsToSpacesFlag(checkCmd)
	addTabsToSpacesFlag(writeCmd)
	addSpacesToTabFlag(checkCmd)
	addSpacesToTabFlag(writeCmd)
	checkCmd.MarkFlagsMutuallyExclusive("replace-tabs-with-spaces", "replace-spaces-with-tabs")
	writeCmd.MarkFlagsMutuallyExclusive("replace-tabs-with-spaces", "replace-spaces-with-tabs")
	checkCmd.SetFlagErrorFunc(stackedFlagErrorFunc)
	writeCmd.SetFlagErrorFunc(stackedFlagErrorFunc)
	rootCmd.SetHelpFunc(rootHelpFunc)
	checkCmd.SetHelpFunc(commandHelpFunc)
	writeCmd.SetHelpFunc(commandHelpFunc)
}

var outputFlagOrder = []string{"silent", "compact", "verbose"}

var configFlagOrder = []string{"line-endings", "replace-tabs-with-spaces", "replace-spaces-with-tabs"}

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
		for _, name := range configFlagOrder {
			if cf := cmd.Flags().Lookup(name); cf != nil {
				printFlagUsage(out, cf, helpTextWidth)
			}
		}
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
	var flagLine string
	if f.Shorthand != "" && f.Name != f.Shorthand {
		flagLine = optionLineIndent + "-" + f.Shorthand + ", --" + f.Name
	} else {
		flagLine = optionLineIndent + "--" + f.Name
	}
	fmt.Fprintln(out, flagLine)
	descCols := width - len(optionDescIndent)
	if descCols < 10 {
		descCols = 10
	}
	usage := wrapWords(f.Usage, descCols)
	for _, line := range strings.Split(usage, "\n") {
		if line == "" {
			fmt.Fprintln(out)
			continue
		}
		fmt.Fprintf(out, "%s%s\n", optionDescIndent, line)
	}
	fmt.Fprintln(out)
}

func optionFlagError(flagName, message string) error {
	return fmt.Errorf("Error:\n%s--%s\n%s%s", optionLineIndent, flagName, optionDescIndent, message)
}

func stackedFlagErrorFunc(_ *cobra.Command, err error) error {
	return fmt.Errorf("Error:\n%s%s", optionDescIndent, err.Error())
}

func formatStderrError(err error) string {
	msg := err.Error()
	if strings.HasPrefix(msg, "Error:\n") {
		return msg
	}
	var b strings.Builder
	b.WriteString("Error:\n")
	for _, line := range strings.Split(msg, "\n") {
		b.WriteString(optionDescIndent)
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", formatStderrError(err))
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
	log.SetOutput(os.Stdout)
	defer log.SetOutput(nil)
	hadIssues, err := Run(true, false, args, rules.LineEndAuto, 0, 0)
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
	tabWidth, err := tabWidthFromCmd(cmd)
	if err != nil {
		return err
	}
	spacesToTab, err := spacesToTabFromCmd(cmd)
	if err != nil {
		return err
	}
	log.SetLevel(outputLevelFromCmd(cmd))
	log.SetOutput(os.Stdout)
	defer log.SetOutput(nil)
	hadIssues, err := Run(true, false, args, mode, tabWidth, spacesToTab)
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
	tabWidth, err := tabWidthFromCmd(cmd)
	if err != nil {
		return err
	}
	spacesToTab, err := spacesToTabFromCmd(cmd)
	if err != nil {
		return err
	}
	log.SetLevel(outputLevelFromCmd(cmd))
	log.SetOutput(os.Stdout)
	defer log.SetOutput(nil)
	_, err = Run(false, true, args, mode, tabWidth, spacesToTab)
	return err
}

func Run(check, doWrite bool, files []string, mode rules.LineEndingMode, tabWidth int, spacesToTab int) (hadIssues bool, err error) {
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
		issues, err := rules.CheckFile(path, mode, tabWidth, spacesToTab)
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
				fmt.Fprintf(os.Stderr, "rules: %s -> %d issue(s): %s\n", path, len(issues), strings.Join(ids, ", "))
			}
		}
	}
	if check {
		if lvl >= log.Normal {
			summaryW := os.Stdout
			if len(allIssues) > 0 {
				summaryW = os.Stderr
			}
			if err := report.WriteSplit(os.Stderr, summaryW, report.FormatCompact, allIssues, len(scanned), scanned); err != nil {
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
		if err := fix.Apply(path, mode, tabWidth, spacesToTab); err != nil {
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
