package commands

import (
	"context"
	"fmt"
	"os"
	"strings"

	docopt "github.com/docopt/docopt-go"

	"e2e_result/commands/show"
)

// Show takes keyword then calls subcommand.
func Show(ctx context.Context, args []string) error {
	doc := `Usage:
	e2e_result show <command> [<args>...]

    results     show e2e automatic tagging test result history.
    runs		show list of available (vcs and ucs) runs for which results were collected.

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
				"Invalid option: 'sveltosctl %s'. Use flag '--help' to read about a specific subcommand.\n",
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
	default:
		fmt.Println(doc)
	}

	return nil
}
