package es_utils

import (
	"context"
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"github.com/olekukonko/tablewriter"
	elastic "github.com/olivere/elastic/v7"
)

func DisplayRuns(ctx context.Context, logger logr.Logger,
	vcs, ucs bool,
	maxResult int,
) error {
	c, err := GetClient(resultCloudstackESURL)
	if err != nil {
		logger.Info(fmt.Sprintf("Failed to get client: %v", err))
		return err
	}

	if err = VerifyIndex(ctx, c, resultCloudstackIndex); err != nil {
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

func GetAvailableRuns(ctx context.Context, c *elastic.Client,
	match string, maxResult int, logger logr.Logger) (*elastic.AggregationBucketKeyItems, error) {
	field := "run"
	termAggr := elastic.NewTermsAggregation().Field(field).Size(maxResult).Order("_key", false)
	searchResult, err := c.Search().Index(resultCloudstackIndex).
		Query(elastic.NewMatchQuery("environment", match)).
		Aggregation(field, termAggr).
		Do(ctx)
	if err != nil {
		logger.Info(fmt.Sprintf("Failed to run query %v", err))
		return nil, err
	}

	logger.Info(fmt.Sprintf("total hits: %v\n\n", searchResult.Hits.TotalHits))

	b, found := searchResult.Aggregations.Terms(field)
	if !found {
		logger.Info("Not found")
		return nil, fmt.Errorf("failed to get term aggregation results")
	}

	return b, nil
}

func aggregationQueryForRun(ctx context.Context, c *elastic.Client,
	match string, maxResult int, table *tablewriter.Table,
	logger logr.Logger) error {
	b, err := GetAvailableRuns(ctx, c, match, maxResult, logger)
	if err != nil {
		return err
	}

	for _, bucket := range b.Buckets {
		table.Append([]string{match, bucket.KeyNumber.String()})
	}
	return nil
}
