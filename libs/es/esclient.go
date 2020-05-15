package es

import (
	"pzatest/config"
	"pzatest/libs/utils"
	"strings"
	"sync"
	"time"

	"github.com/olivere/elastic"
	"github.com/op/go-logging"
)

var (
	logger             = logging.MustGetLogger("test")
	cachedESClients    = make(map[string]*elastic.Client, 10)
	cacheESClientsSync sync.RWMutex
)

// NewESClient build the ES client
func NewESClient(user, password string, urls ...string) (*elastic.Client, error) {
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

	return client, nil
}

// Index ...
func Index(c *elastic.Client) error {
	return nil
}
