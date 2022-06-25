package commands

import (
	"context"
	"fmt"
	"os"
	"strings"

	docopt "github.com/docopt/docopt-go"

	"github.com/gianlucam76/cs-e2e-result/commands/show"
)

// Show takes keyword then calls subcommand.
func Show(ctx context.Context, args []string) error {
	doc := `Usage:
	e2e_result show <command> [<args>...]

    results     show e2e automatic tagging test result history.
    runs        show list of available (vcs and ucs) runs for which results were collected.
    reports     show e2e reports.
    usage       show e2e usage reports.

Options:
	-h --help      Show this screen.

Description:
	See 'e2e_result show <command> --help' to read about a specific subcommand.
  `

	parser := &docopt.Parser{
		HelpHandler:   docopt.PrintHelpAndExit,
		OptionsFirst:  true,
		SkipHelpFlags: false,
	}

	opts, err := parser.ParseArgs(doc, nil, "1.0")
	if err != nil {
		if _, ok := err.(*docopt.UserError); ok {
			fmt.Printf(
				"Invalid option: 'e2e_result %s'. Use flag '--help' to read about a specific subcommand.\n",
				strings.Join(os.Args[1:], " "),
			)
		}
		os.Exit(1)
	}

	command := opts["<command>"].(string)
	arguments := append([]string{"show", command}, opts["<args>"].([]string)...)

	switch command {
	case "results":
		return show.ResultHistory(ctx, arguments)
	case "runs":
		return show.AvailableRuns(ctx, arguments)
	case "reports":
		return show.ReportHistory(ctx, arguments)
	case "usage":
		return show.UsageHistory(ctx, arguments)
	default:
		fmt.Println(doc)
	}

	return nil
}
