package es_utils

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/olekukonko/tablewriter"
	elastic "github.com/olivere/elastic/v7"
)

const (
	usageCloudstackESURL = "http://172.31.165.56:9200"
	usageCloudstackIndex = "cs_e2e_usage_entries"
)

type UsageReport struct {
	// Name identifies the pod this usage report is about.
	// If pod is part of a deployment, use <namespace>/<deployment name>
	// If pod is part of a daemonset, use <namespace>/<daemonset name>
	// Otherwise, use <namespace>/<pod name>
	Name string `json:"name"`
	// Memory is the max memory usage seen in Ki
	Memory int64 `json:"memory"`
	// CPU is the max CPU usage seen in m
	CPU int64 `json:"cpu"`
	// MemoryLimit is the pod memory limit in Ki
	MemoryLimit int64 `json:"memoryLimit"`
	// CPULimit is the pod CPU limit in m
	CPULimit int64 `json:"cpuLimit"`
	// Environment represents the environment where e2e ran, i.e UCS or VCS
	Environment string `json:"environment"`
	// Run is the sanity run id
	Run int `json:"run"`
	// CreatedTime is the time entry was created
	CreatedTime time.Time `json:"createdTime"`
}

func GetUsageReports(ctx context.Context, logger logr.Logger,
	run, pod string,
	vcs, ucs bool, maxResult int,
) (*elastic.SearchResult, error) {
	c, err := GetClient(usageCloudstackESURL)
	if err != nil {
		logger.Info(fmt.Sprintf("Failed to get client: %v", err))
		return nil, err
	}

	if err = VerifyIndex(ctx, c, usageCloudstackIndex); err != nil {
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

	if pod != "" {
		logger.Info(fmt.Sprintf("Filter by report name:%s", pod))
		generalQ.Filter(elastic.NewTermQuery("name.keyword", pod)) // Exact match
	}

	searchResult, err := c.Search().Index(usageCloudstackIndex).Query(generalQ).Size(maxResult).
		SortBy(elastic.NewFieldSort("run").Desc().SortMode("max")).
		Pretty(true).            // pretty print request and response JSON
		Do(context.Background()) // execute
	if err != nil {
		logger.Info(fmt.Sprintf("Failed to run query %v", err))
		return nil, err
	}

	logger.Info(fmt.Sprintf("Query took %d milliseconds\n", searchResult.TookInMillis))

	return searchResult, nil
}

func DisplayUsageReport(ctx context.Context, logger logr.Logger,
	run, pod, usageType string,
	vcs, ucs bool,
	maxResult int,
) error {
	searchResult, err := GetUsageReports(ctx, logger, run, pod, vcs, ucs, maxResult)
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ENVIRONMENT", "RUN", "POD NAME", "TYPE", "MAX USED", "LIMIT"})
	table.SetAutoWrapText(false)
	table.SetRowLine(true)

	var rtyp UsageReport
	for _, item := range searchResult.Each(reflect.TypeOf(rtyp)) {
		r := item.(UsageReport)
		if usageType == "" || strings.EqualFold(usageType, "memory") {
			table.Append([]string{r.Environment, strconv.Itoa(r.Run),
				r.Name, "Memory", fmt.Sprintf("%dKi", r.Memory), fmt.Sprintf("%dKi", r.MemoryLimit)})
		}
		if usageType == "" || strings.EqualFold(usageType, "cpu") {
			table.Append([]string{r.Environment, strconv.Itoa(r.Run),
				r.Name, "CPU", fmt.Sprintf("%dm", r.CPU), fmt.Sprintf("%dm", r.CPULimit)})
		}
	}
	table.Render()

	return nil
}
