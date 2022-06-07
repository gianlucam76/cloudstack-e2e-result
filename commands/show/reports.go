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

// ReportHistory displays information about e2e sanity entries.
func ReportHistory(ctx context.Context, args []string) error {
	doc := `Usage:
	e2e_result show reports [--vcs | --ucs] [--run=<id>] [--type=<name>] [--subtype=<name>] [--name=<name>] [--max=<int>]
Options:
  -h --help               Show this screen.
     --vcs                Show e2e test results in vcs run.
     --ucs                Show e2e test results in ucs run.
     --run=<id>           Show e2e test results in a specific (vcs or ucs) run 
     --max=<int>          Maximum number of results to display (default is 100)
     --type=<name>        Show history for a report type.
     --sybtype=<name>     Show history for a report subtype.
     --name=<name>        Show history of a specific reports.

Description:
  The show reports command shows information about e2e reports.
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

	reportType := ""
	if passedReportType := parsedArgs["--type"]; passedReportType != nil {
		reportType = passedReportType.(string)
	}

	reportSubType := ""
	if passedReportSubType := parsedArgs["--subtype"]; passedReportSubType != nil {
		reportSubType = passedReportSubType.(string)
	}

	reportName := ""
	if passedReportName := parsedArgs["--name"]; passedReportName != nil {
		reportName = passedReportName.(string)
	}

	max := 100
	if passedMax := parsedArgs["--max"]; passedMax != nil {
		max, err = strconv.Atoi(passedMax.(string))
		if err != nil {
			return err
		}
	}

	return es_utils.DisplayReport(context.TODO(), logger, run, reportType, reportSubType, reportName, vcs, ucs, max)
}
