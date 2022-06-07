package es_utils

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	"github.com/olekukonko/tablewriter"
	elastic "github.com/olivere/elastic/v7"
)

const (
	resultCloudstackESURL = "http://172.31.165.56:9200"
	resultCloudstackIndex = "cs_e2e"
)

type Result struct {
	// Name is the name of the test
	Name string `json:"name"`
	// Description is the test description
	Description string `json:"description"`
	// Maintainer is the maintainer for a given test
	Maintainer string `json:"maintainer"`
	// DurationInMinutes is the duration of the test in minutes
	DurationInMinutes float64 `json:"durationInMinutes"`
	// Duration is the duration of the test in seconds
	DurationInSecond time.Duration `json:"durationInSeconds"`
	// Result indicates whether test passed or failed or it was skipped
	Result string `json:"result"`
	// Environment represents the environment where e2e ran, i.e UCS or VCS
	Environment string `json:"environment"`
	// Run is the sanity run id
	Run int `json:"run"`
	// StartTime is the time test started
	StartTime time.Time `json:"startTime"`
	// Serial indicates whether test was run in serial
	Serial bool `json:"serial"`
}

func GetResults(ctx context.Context, logger logr.Logger,
	run, testName string,
	vcs, ucs, passed, failed, skipped bool,
	maxResult int,
) (*elastic.SearchResult, error) {
	c, err := GetClient(resultCloudstackESURL)
	if err != nil {
		logger.Info(fmt.Sprintf("Failed to get client: %v", err))
		return nil, err
	}

	if err = VerifyIndex(ctx, c, resultCloudstackIndex); err != nil {
		logger.Info(fmt.Sprintf("Failed to verify index %v", err))
		return nil, err
	}

	generalQ := elastic.NewBoolQuery().Should()

	if passed {
		logger.Info("Filter by result:passed")
		generalQ.Filter(elastic.NewMatchQuery("result", "passed"))
	} else if failed {
		logger.Info("Filter by result:failed")
		generalQ.Filter(elastic.NewMatchQuery("result", "failed"))
	} else if skipped {
		logger.Info("Filter by result:skipped")
		generalQ.Filter(elastic.NewMatchQuery("result", "skipped"))
	}

	if vcs {
		logger.Info("Filter by environment:vcs")
		generalQ.Filter(elastic.NewMatchQuery("environment", "vcs"))
	} else if ucs {
		logger.Info("Filter by environment:ucs")
		generalQ.Filter(elastic.NewMatchQuery("environment", "ucs"))
	}

	if run != "" {
		logger.Info(fmt.Sprintf("Filter by run:%s", run))
		generalQ.Filter(elastic.NewMatchQuery("run", run))
	}

	if testName != "" {
		logger.Info(fmt.Sprintf("Filter by test:%s", testName))
		generalQ.Filter(elastic.NewTermQuery("name.keyword", testName)) // Exact match
	}

	searchResult, err := c.Search().Index(resultCloudstackIndex).Query(generalQ).Size(maxResult).
		Pretty(true).            // pretty print request and response JSON
		Do(context.Background()) // execute
	if err != nil {
		logger.Info(fmt.Sprintf("Failed to run query %v", err))
		return nil, err
	}

	logger.Info(fmt.Sprintf("Query took %d milliseconds\n", searchResult.TookInMillis))

	return searchResult, nil
}

func DisplayResult(ctx context.Context, logger logr.Logger,
	run, testName string,
	vcs, ucs, passed, failed, skipped bool,
	maxResult int,
) error {
	searchResult, err := GetResults(ctx, logger, run, testName, vcs, ucs, passed, failed, skipped, maxResult)
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ENVIRONMENT", "RUN", "TEST", "RESULT", "DURATION"})
	table.SetAutoWrapText(false)
	table.SetRowLine(true)

	var rtyp Result
	for _, item := range searchResult.Each(reflect.TypeOf(rtyp)) {
		r := item.(Result)
		name := r.Name
		if r.Serial {
			name = fmt.Sprintf("%s*", r.Name)
		}
		table.Append([]string{r.Environment, strconv.Itoa(r.Run), name,
			r.Result, fmt.Sprintf("%f", r.DurationInMinutes)})
	}

	table.Render()

	return nil
}
