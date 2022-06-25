package show

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	docopt "github.com/docopt/docopt-go"
	"k8s.io/klog/v2/klogr"

	"github.com/gianlucam76/cs-e2e-result/es_utils"
)

// UsageHistory displays information about e2e sanity usage entries.
func UsageHistory(ctx context.Context, args []string) error {
	doc := `Usage:
	e2e_result show usage [--vcs | --ucs] [--run=<id>] [--pod=<name>] [--type=<type>] [--max=<int>]
Options:
  -h --help               Show this screen.
     --vcs                Show e2e test results in vcs run.
     --ucs                Show e2e test results in ucs run.
     --run=<id>           Show e2e test results in a specific (vcs or ucs) run 
     --max=<int>          Maximum number of results to display (default is 100)
     --pod=<name>         Show history of a specific pod usage.
     --type=<type>        Memory or CPU

Description:
  The show usage command shows information about e2e usage reports.
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

	run := ""
	if passedRun := parsedArgs["--run"]; passedRun != nil {
		run = passedRun.(string)
	}

	podName := ""
	if passedPodName := parsedArgs["--pod"]; passedPodName != nil {
		podName = passedPodName.(string)
	}

	usageType := ""
	if passedUsageType := parsedArgs["--type"]; passedUsageType != nil {
		usageType = passedUsageType.(string)
	}

	max := 100
	if passedMax := parsedArgs["--max"]; passedMax != nil {
		max, err = strconv.Atoi(passedMax.(string))
		if err != nil {
			return err
		}
	}

	return es_utils.DisplayUsageReport(context.TODO(), logger, run, podName, usageType, vcs, ucs, max)
}
