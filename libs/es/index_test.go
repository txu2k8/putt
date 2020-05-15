package es

import (
	"testing"
)

func TestIndex(t *testing.T) {
	indexNameArr := []string{"test1", "test2", "test3"}
	client, _ := NewESClient("root", "password", "http://10.180.128.11:30380")
	for _, indexName := range indexNameArr {
		bulk := Bulker{
			c:           client,
			workers:     3,
			index:       indexName,
			bulkActions: 1000,
			bulkSize:    4096,
		}
		bulk.Run()
	}

}
