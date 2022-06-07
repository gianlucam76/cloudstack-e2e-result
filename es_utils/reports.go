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
	reportCloudstackESURL = "http://172.31.165.56:9200"
	reportCloudstackIndex = "cs_e2e_entries"
)

type Report struct {
	// Type of the report
	Type string `json:"type"`
	// Name is the name of the instance this report is about
	Name string `json:"name"`
	// SubType is an optional field.
	// It can be used to further specify a type of a report.
	// If set, reports in the ReportType categories, will be aggregate by
	// SubType field.
	// For instance, when reporting a cluster is ready (all machines are up
	// and features deployed) clusters with different number of nodes will
	// have significantly different durations. So SubType will be set to
	// a string representing the number of nodes in the cluster.
	SubType string `json:"subType"`
	// Duration is the time taken in minutes
	DurationInMinutes float64 `json:"durationInMinutes"`
	// Environment represents the environment where e2e ran, i.e UCS or VCS
	Environment string `json:"environment"`
	// Run is the sanity run id
	Run int `json:"run"`
	// CreatedTime is the time entry was created
	CreatedTime time.Time `json:"createdTime"`
}

func GetReports(ctx context.Context, logger logr.Logger,
	run, reportType, reportSubType, reportName string,
	vcs, ucs bool, maxResult int,
) (*elastic.SearchResult, error) {
	c, err := GetClient(reportCloudstackESURL)
	if err != nil {
		logger.Info(fmt.Sprintf("Failed to get client: %v", err))
		return nil, err
	}

	if err = VerifyIndex(ctx, c, reportCloudstackIndex); err != nil {
		logger.Info(fmt.Sprintf("Failed to verify index %v", err))
		return nil, err
	}

	generalQ := elastic.NewBoolQuery().Should()

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

	if reportType != "" {
		logger.Info(fmt.Sprintf("Filter by reportType:%s", reportType))
		generalQ.Filter(elastic.NewMatchQuery("type", reportType))
	}

	if reportSubType != "" {
		logger.Info(fmt.Sprintf("Filter by reportSubType:%s", reportSubType))
		generalQ.Filter(elastic.NewTermQuery("subType.keyword", reportSubType))
	}

	if reportName != "" {
		logger.Info(fmt.Sprintf("Filter by report name:%s", reportName))
		generalQ.Filter(elastic.NewTermQuery("name.keyword", reportName)) // Exact match
	}

	searchResult, err := c.Search().Index(reportCloudstackIndex).Query(generalQ).Size(maxResult).
		Pretty(true).            // pretty print request and response JSON
		Do(context.Background()) // execute
	if err != nil {
		logger.Info(fmt.Sprintf("Failed to run query %v", err))
		return nil, err
	}

	logger.Info(fmt.Sprintf("Query took %d milliseconds\n", searchResult.TookInMillis))

	return searchResult, nil
}

func DisplayReport(ctx context.Context, logger logr.Logger,
	run, reportType, reportSubType, reportName string,
	vcs, ucs bool,
	maxResult int,
) error {
	searchResult, err := GetReports(ctx, logger, run, reportType, reportSubType, reportName, vcs, ucs, maxResult)
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ENVIRONMENT", "RUN", "REPORT TYPE", "REPORT SUBTYPE", "NAME", "DURATION"})
	table.SetAutoWrapText(false)
	table.SetRowLine(true)

	var rtyp Report
	for _, item := range searchResult.Each(reflect.TypeOf(rtyp)) {
		r := item.(Report)
		table.Append([]string{r.Environment, strconv.Itoa(r.Run),
			r.Type, r.SubType, r.Name, fmt.Sprintf("%f", r.DurationInMinutes)})
	}

	table.Render()

	return nil
}
