package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"

	docopt "github.com/docopt/docopt-go"

	"github.com/gianlucam76/cs-e2e-result/commands"
)

func main() {
	ctx := context.Background()

	klog.InitFlags(nil)
	logger := klogr.New()

	logger.Info("e2e_result tool")

	doc := `Usage:
	e2e_result [options] <command> [<args>...]

	show          Display information on e2e results

Options:
  -h --help     Show this screen.

Description:
  The e2e_result command line tool is used to display e2e results.
  See 'e2e_result <command> --help' to read about a specific subcommand.
`

	parser := &docopt.Parser{
		HelpHandler:   docopt.PrintHelpOnly,
		OptionsFirst:  true,
		SkipHelpFlags: false,
	}

	opts, err := parser.ParseArgs(doc, nil, "")
	if err != nil {
		if _, ok := err.(*docopt.UserError); ok {
			fmt.Printf(
				"Invalid option: 'e2e_result %s'. Use flag '--help' to read about a specific subcommand.\n",
				strings.Join(os.Args[1:], " "),
			)
		}
		os.Exit(1)
	}

	if opts["<command>"] != nil {
		command := opts["<command>"].(string)
		args := append([]string{command}, opts["<args>"].([]string)...)
		var err error

		switch command {
		case "show":
			err = commands.Show(ctx, args)
		default:
			err = fmt.Errorf("unknown command: %q\n%s", command, doc)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
	}
}
