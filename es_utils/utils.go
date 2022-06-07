package es_utils

import (
	"context"
	"fmt"
	"time"

	elastic "github.com/olivere/elastic/v7"
)

const (
	healthCheckInterval = 10 * time.Second
)

// GetClient returns elastic client
func GetClient(esURL string) (*elastic.Client, error) {
	return elastic.NewClient(
		elastic.SetSniff(false),
		elastic.SetURL(esURL),
		elastic.SetHealthcheckInterval(healthCheckInterval),
	)
}

func VerifyIndex(ctx context.Context, c *elastic.Client, index string) error {
	exists, err := c.IndexExists(index).Do(ctx)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("%s index does not exist", index)
	}

	return nil
}
