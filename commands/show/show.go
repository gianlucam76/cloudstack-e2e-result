package show

import (
	"context"
	"e2e_result/es_utils"
	"fmt"
	"strconv"
	"strings"

	docopt "github.com/docopt/docopt-go"
	"k8s.io/klog/v2/klogr"
)

// ResultHistory displays information about e2e sanity results.
func ResultHistory(ctx context.Context, args []string) error {
	doc := `Usage:
	e2e_result show results [--vcs | --ucs] [--failed | --passed | --skipped] [--run=<id>] [--test=<name>] [--max=<int>]
Options:
  -h --help               Show this screen.
     --vcs                Show e2e test results in vcs run.
     --ucs                Show e2e test results in ucs run.
     --failed             Show e2e test results filtering by failed tests.
     --passed             Show e2e test results filtering by passed tests.
     --skipped            Show e2e test results filtering by skipped tests.
     --run=<id>           Show e2e test results in a specific (vcs or ucs) run 
     --test=<name>        Show history for a specific test.
	 --max=<int>          Maximum number of results to display (default is 100)

Description:
  The show results command shows information about e2e results.
`
	parsedArgs, err := docopt.ParseArgs(doc, nil, "1.0")
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf(
			"invalid option: 'e2e_result %s'. Use flag '--help' to read about a specific subcommand. Error: %v",
			strings.Join(args, " "),
			err,
		)
	}
	if len(parsedArgs) == 0 {
		return nil
	}

	logger := klogr.New()

	vcs := parsedArgs["--vcs"].(bool)
	ucs := parsedArgs["--ucs"].(bool)

	failed := parsedArgs["--failed"].(bool)
	passed := parsedArgs["--passed"].(bool)
	skipped := parsedArgs["--skipped"].(bool)

	run := ""
	if passedRun := parsedArgs["--run"]; passedRun != nil {
		run = passedRun.(string)
	}

	test := ""
	if passedTest := parsedArgs["--test"]; passedTest != nil {
		test = passedTest.(string)
	}

	max := 100
	if passedMax := parsedArgs["--max"]; passedMax != nil {
		max, err = strconv.Atoi(passedMax.(string))
		if err != nil {
			return err
		}
	}

	return es_utils.DisplayResult(context.TODO(), logger, run, test, vcs, ucs, passed, failed, skipped, max)
}
