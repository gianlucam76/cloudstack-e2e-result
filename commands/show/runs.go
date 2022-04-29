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

// AvailableRuns displays information about (vcs and ucs) runs for which results were collected.
func AvailableRuns(ctx context.Context, args []string) error {
	doc := `Usage:
	e2e_result show runs [--vcs | --ucs] [--max=<int>]
Options:
  -h --help               Show this screen.
     --vcs                Show e2e test results in vcs run.
     --ucs                Show e2e test results in ucs run.
     --max=<int>          Maximum number of results to display (default is 100)

Description:
  The show runs command shows information about available runs for which results were collected.
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

	max := 100
	if passedMax := parsedArgs["--max"]; passedMax != nil {
		max, err = strconv.Atoi(passedMax.(string))
		if err != nil {
			return err
		}
	}

	return es_utils.DisplayRuns(context.TODO(), logger, ucs, vcs, max)
}
