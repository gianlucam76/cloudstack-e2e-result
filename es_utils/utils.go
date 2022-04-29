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
	cloudstackESURL     = "http://172.31.165.56:9200"
	cloudstackIndex     = "cs_e2e"
	healthCheckInterval = 10 * time.Second
)

type result struct {
	// Name is the name of the test
	Name string `json:"name"`
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
}

// getClient returns elastic client
func getClient() (*elastic.Client, error) {
	return elastic.NewClient(
		elastic.SetSniff(false),
		elastic.SetURL(cloudstackESURL),
		elastic.SetHealthcheckInterval(healthCheckInterval),
	)
}

func verifyIndex(ctx context.Context, c *elastic.Client) error {
	exists, err := c.IndexExists(cloudstackIndex).Do(ctx)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("%s index does not exist", cloudstackIndex)
	}

	return nil
}

func DisplayResult(ctx context.Context, logger logr.Logger,
	run, testName string,
	vcs, ucs, passed, failed, skipped bool,
	maxResult int,
) error {
	c, err := getClient()
	if err != nil {
		logger.Info(fmt.Sprintf("Failed to get client: %v", err))
		return err
	}

	if err = verifyIndex(ctx, c); err != nil {
		logger.Info(fmt.Sprintf("Failed to verify index %v", err))
		return err
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
		generalQ.Filter(elastic.NewMatchQuery("name", testName))
	}

	searchResult, err := c.Search().Index(cloudstackIndex).Query(generalQ).Size(maxResult).
		Pretty(true).            // pretty print request and response JSON
		Do(context.Background()) // execute
	if err != nil {
		logger.Info(fmt.Sprintf("Failed to run query %v", err))
		return err
	}

	logger.Info(fmt.Sprintf("Query took %d milliseconds\n", searchResult.TookInMillis))

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ENVIRONMENT", "RUN", "TEST", "RESULT", "DURATION"})
	table.SetAutoWrapText(false)
	table.SetRowLine(true)

	var rtyp result
	for _, item := range searchResult.Each(reflect.TypeOf(rtyp)) {
		r := item.(result)
		table.Append([]string{r.Environment, strconv.Itoa(r.Run), r.Name,
			r.Result, fmt.Sprintf("%f", r.DurationInMinutes)})
	}

	table.Render()

	return nil
}

func DisplayRuns(ctx context.Context, logger logr.Logger,
	vcs, ucs bool,
	maxResult int,
) error {
	c, err := getClient()
	if err != nil {
		logger.Info(fmt.Sprintf("Failed to get client: %v", err))
		return err
	}

	if err = verifyIndex(ctx, c); err != nil {
		logger.Info(fmt.Sprintf("Failed to verify index %v", err))
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ENVIRONMENT", "RUN"})
	table.SetAutoWrapText(false)
	table.SetRowLine(true)

	if vcs || (!vcs && !ucs) {
		if err := aggregationQueryForRun(ctx, c, "vcs", maxResult, table, logger); err != nil {
			return err
		}
	}
	if ucs || (!vcs && !ucs) {
		if err := aggregationQueryForRun(ctx, c, "ucs", maxResult, table, logger); err != nil {
			return err
		}
	}

	table.Render()

	return nil
}

func aggregationQueryForRun(ctx context.Context, c *elastic.Client,
	match string, maxResult int, table *tablewriter.Table,
	logger logr.Logger) error {
	termAggr := elastic.NewTermsAggregation().Field("run")
	searchResult, err := c.Search().Index(cloudstackIndex).
		Query(elastic.NewMatchQuery("environment", match)).
		Aggregation("run", termAggr).
		Size(maxResult).Do(ctx)
	if err != nil {
		logger.Info(fmt.Sprintf("Failed to run query %v", err))
		return err
	}

	b, found := searchResult.Aggregations.Terms("run")
	if !found {
		logger.Info("Not found")
		return nil
	}

	for _, bucket := range b.Buckets {
		table.Append([]string{match, bucket.KeyNumber.String()})
	}
	return nil
}
