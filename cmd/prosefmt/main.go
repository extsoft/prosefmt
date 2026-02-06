package prosefmt

import (
	"fmt"
	"io"
	"os"
	"prosefmt/internal/fix"
	"prosefmt/internal/log"
	"prosefmt/internal/report"
	"prosefmt/internal/rules"
	"prosefmt/internal/scanner"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	version        = "dev"
	checkFlag      bool
	writeFlag      bool
	silentFlag     bool
	compactFlag    bool
	verboseFlag    bool
	checkHadIssues bool
)

const rootDescription = "The simplest text formatter for making your files look correct."

var rootCmd = &cobra.Command{
	Use:   "prosefmt [command] [flags] [path...]",
	Short: rootDescription,
	Long:  rootDescription,
	Args:  cobra.ArbitraryArgs,
	RunE:  runE,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.Flags().BoolVar(&checkFlag, "check", false, "check files and report issues (default)")
	rootCmd.Flags().BoolVar(&writeFlag, "write", false, "write fixes in place")
	rootCmd.PersistentFlags().BoolVar(&silentFlag, "silent", false, "no standard output printed")
	rootCmd.PersistentFlags().BoolVar(&compactFlag, "compact", false, "show formatted or errored files (default)")
	rootCmd.PersistentFlags().BoolVar(&verboseFlag, "verbose", false, "print debug output (steps, scanner, rules, timing)")
	rootCmd.SetHelpFunc(helpFunc)
}

var optionFlagNames = map[string]bool{"check": true, "write": true}
var outputFlagOrder = []string{"silent", "compact", "verbose"}
var outputFlagNames = map[string]bool{"silent": true, "compact": true, "verbose": true}

func helpFunc(cmd *cobra.Command, args []string) {
	out := cmd.OutOrStderr()
	if cmd.Short != "" {
		fmt.Fprintf(out, "%s\n\n", cmd.Short)
	}
	fmt.Fprintf(out, "Usage:\n  %s\n\n", cmd.UseLine())
	if len(cmd.Commands()) > 0 {
		fmt.Fprintln(out, "Commands:")
		for _, c := range cmd.Commands() {
			if c.IsAvailableCommand() {
				fmt.Fprintf(out, "  %s\t%s\n", c.Name(), c.Short)
			}
		}
		fmt.Fprintln(out, "")
	}
	fmt.Fprintln(out, "Options:")
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if optionFlagNames[f.Name] {
			printFlagUsage(out, f)
		}
	})
	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		if optionFlagNames[f.Name] {
			printFlagUsage(out, f)
		}
	})
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Output:")
	for _, name := range outputFlagOrder {
		if f := cmd.PersistentFlags().Lookup(name); f != nil {
			printFlagUsage(out, f)
		}
	}
	if version != "" {
		fmt.Fprintf(out, "\nVersion: %s\n", version)
	}
}

func printFlagUsage(out io.Writer, f *pflag.Flag) {
	if f.Shorthand != "" && f.Name != f.Shorthand {
		fmt.Fprintf(out, "  -%s, --%s\t%s\n", f.Shorthand, f.Name, f.Usage)
	} else {
		fmt.Fprintf(out, "      --%s\t%s\n", f.Name, f.Usage)
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

func verbosityLevel() log.Level {
	if silentFlag {
		return log.Silent
	}
	if verboseFlag {
		return log.Verbose
	}
	if compactFlag {
		return log.Normal
	}
	return log.Normal
}

func runE(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		helpFunc(cmd, nil)
		return nil
	}
	if checkFlag && writeFlag {
		return fmt.Errorf("cannot use both --check and --write")
	}
	if !checkFlag && !writeFlag {
		checkFlag = true
	}
	log.SetLevel(verbosityLevel())
	hadIssues, err := run(checkFlag, writeFlag, args)
	if err != nil {
		return err
	}
	checkHadIssues = checkFlag && hadIssues
	return nil
}

func run(check, doWrite bool, paths []string) (hadIssues bool, err error) {
	start := time.Now()
	lvl := log.GetLevel()
	if lvl >= log.Verbose {
		log.Logf(log.Verbose, "Configuration: check=%v paths=%v\n", check, paths)
	}
	files, skipped, err := scanner.Scan(paths)
	if err != nil {
		return false, err
	}
	elapsedScan := time.Since(start)
	if lvl >= log.Verbose {
		if len(files) == 0 {
			log.Logf(log.Verbose, "No text files found. Scanned 0 text file(s), skipped %d path(s).\n", len(skipped))
		} else {
			log.Logf(log.Verbose, "Scanned %d text file(s), skipped %d path(s).\n", len(files), len(skipped))
		}
		for _, p := range sortedKeys(skipped) {
			log.Logf(log.Verbose, "scanner: rejected %s (reason: %s)\n", p, skipped[p])
		}
		for _, p := range files {
			log.Logf(log.Verbose, "scanner: accepted %s\n", p)
		}
	}
	if len(files) == 0 {
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
	for _, path := range files {
		if lvl >= log.Verbose {
			if check {
				log.Logf(log.Verbose, "Checking %s\n", path)
			} else {
				log.Logf(log.Verbose, "Writing %s\n", path)
			}
		}
		issues, err := rules.CheckFile(path)
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
			if err := report.Write(os.Stdout, report.FormatCompact, allIssues, len(files), files); err != nil {
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
		if err := fix.Apply(path); err != nil {
			return false, err
		}
		if lvl >= log.Verbose {
			log.Logf(log.Verbose, "write: applied to %s\n", path)
		}
	}
	if lvl >= log.Normal && len(fileIssues) > 0 {
		paths := make([]string, 0, len(fileIssues))
		for p := range fileIssues {
			paths = append(paths, p)
		}
		sort.Strings(paths)
		fmt.Fprintf(os.Stdout, "Wrote %d file(s):\n", len(paths))
		for _, p := range paths {
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
