package es

import (
	"gtest/config"
	"index/utils"
	"strings"
	"sync"
	"time"

	"github.com/olivere/elastic"
	"learngo/src/github.com/golang/glog"
)

var (
	cachedESClients    = make(map[string]*elastic.Client, 10)
	cacheESClientsSync sync.RWMutex
)

// NewESClient build the ES client
func NewESClient(user, password string, urls ...string) (*elastic.Client, error) {

	defer utils.TimeTrack(time.Now(), "NewESClient")

	glog.V(4).Infof("Create elasticsearch client, user %v, urls %v", user, urls)

	cacheKey := strings.Join(urls, "-")

	cacheESClientsSync.RLock()
	client, hit := cachedESClients[cacheKey]
	cacheESClientsSync.RUnlock()

	if !hit {
		glog.V(4).Infof("Miss elasticsearch client cache, will create new client")

		var err error
		client, err = elastic.NewClient(
			elastic.SetURL(urls...),
			elastic.SetBasicAuth(user, password),
			elastic.SetSniff(false),
			elastic.SetHealthcheckInterval(time.Second*time.Duration(config.Config.ES.HealthcheckInterval)),
		)
		if err != nil {
			glog.Errorf("Connect to the Elasticsearch server error: %v", err)
			return nil, err
		}

		cacheESClientsSync.Lock()
		defer cacheESClientsSync.Unlock()
		cachedESClients[cacheKey] = client
	}

	return client, nil
}
