package es

import (
	"context"
	"putt/config"
	"putt/libs/utils"
	"strings"
	"sync"
	"time"

	"github.com/olivere/elastic/v6"
	"github.com/op/go-logging"
)

// DOC ...
const (
	DOC = "doc"
)

var (
	logger             = logging.MustGetLogger("test")
	cachedESClients    = make(map[string]*elastic.Client, 10)
	cacheESClientsSync sync.RWMutex
)

// NewESClient build the ES client
func NewESClient(user, password string, urls ...string) (*elastic.Client, error) {
	ctx := context.Background()

	defer utils.TimeTrack(time.Now(), "NewESClient")
	logger.Infof("Create elasticsearch client, user %v, urls %v", user, urls)
	cacheKey := strings.Join(urls, "-")
	cacheESClientsSync.RLock()
	client, hit := cachedESClients[cacheKey]
	cacheESClientsSync.RUnlock()
	if !hit {
		logger.Infof("Miss elasticsearch client cache, will create new client")
		var err error
		client, err = elastic.NewClient(
			elastic.SetURL(urls...),
			elastic.SetBasicAuth(user, password),
			elastic.SetSniff(false),
			elastic.SetHealthcheckInterval(time.Second*time.Duration(config.Config.ES.HealthcheckInterval)),
		)
		if err != nil {
			logger.Errorf("Connect to the Elasticsearch server error: %v", err)
			return nil, err
		}

		cacheESClientsSync.Lock()
		defer cacheESClientsSync.Unlock()
		cachedESClients[cacheKey] = client
	}

	// Ping the Elasticsearch server to get e.g. the version number
	info, code, err := client.Ping(urls[0]).Do(ctx)
	if err != nil {
		// Handle error
		panic(err)
	}
	logger.Infof("Elasticsearch returned with code %d and version %s\n", code, info.Version.Number)

	// Getting the ES version number is quite common, so there's a shortcut
	esversion, err := client.ElasticsearchVersion(urls[0])
	if err != nil {
		// Handle error
		panic(err)
	}
	logger.Infof("Elasticsearch version %s\n", esversion)

	return client, nil
}

// CreateIndex ...
// body: The configuration for the index (`settings` and `mappings`)
func CreateIndex(client *elastic.Client, indexName string, body string) error {
	ctx := context.Background()
	// Use the IndexExists service to check if a specified index exists.
	exists, err := client.IndexExists(indexName).Do(ctx)
	if err != nil {
		// Handle error
		panic(err)
	}
	if exists {
		logger.Infof("Index with name %s already exists, Skip create!", indexName)
	} else {
		// Create a new index.
		createIndex, err := client.CreateIndex(indexName).BodyString(body).Do(ctx)
		if err != nil {
			// Handle error
			panic(err)
		}
		if !createIndex.Acknowledged {
			// Not acknowledged
		}
	}

	return nil
}

// DeleteIndex ...
func DeleteIndex(client *elastic.Client, indexName string) error {
	ctx := context.Background()
	exists, err := client.IndexExists(indexName).Do(ctx)
	if err != nil {
		// Handle error
		panic(err)
	}
	if exists {
		deleteIndex, err := client.DeleteIndex(indexName).Do(ctx)
		if err != nil {
			// Handle error
			panic(err)
		}
		if !deleteIndex.Acknowledged {
			// Not acknowledged
		}
	}
	return nil
}

// BulkIndex ...
func BulkIndex(esClient *elastic.Client, indexName string, dataList *[]map[string]interface{}, mode string) (bool, error) {
	bulkProcessor, err := esClient.BulkProcessor().
		Workers(4).
		BulkActions(1000).
		BulkSize(15 << 20).
		Stats(true).
		Do(context.Background())
	if err != nil {
		logger.Errorf("failed to create bulk processor: %v", err)
		return false, err
	}

	defer bulkProcessor.Close()

	for _, dataMap := range *dataList {
		if dataMap != nil {
			switch mode {
			case "add":
				if id, hit := dataMap["id"]; hit {
					bulkProcessor.Add(
						elastic.NewBulkIndexRequest().
							Index(indexName).
							Type(DOC).
							Id(id.(string)).
							Doc(dataMap["data"]),
					)
				} else {
					// todo test not pass
					bulkProcessor.Add(
						elastic.NewBulkIndexRequest().
							Index(indexName).
							Type(DOC).
							Doc(dataMap["data"]),
					)
				}
			case "update":
				bulkProcessor.Add(
					elastic.NewBulkUpdateRequest().
						Index(indexName).
						Type(DOC).
						Id(dataMap["id"].(string)).
						Doc(dataMap["data"]),
				)
			case "delete":
				bulkProcessor.Add(
					elastic.NewBulkDeleteRequest().
						Index(indexName).
						Type(DOC).
						Id(dataMap["id"].(string)),
				)
			}
		}
	}

	err = bulkProcessor.Start(context.Background())

	if err != nil {
		return false, err
	}

	return true, nil
}
