package es

import (
	"fmt"
	"pzatest/libs/utils"
)

// DocType ...
const DocType = "doc"

// CreateIndexBody : for create index body, settings && mappings
const CreateIndexBody = `
{
	"settings":{
        "analysis": {
            "analyzer": {
                "test_analyzer": {
                    "type": "custom",
                    "tokenizer": "keyword",
                    "filter": [
                        "lowercase"
                    ]
                }
            }
        }
        "index.indexing.slowlog.threshold.index.debug": "2s",
        "index.indexing.slowlog.threshold.index.info": "5s",
        "index.indexing.slowlog.threshold.index.trace": "500ms",
        "index.indexing.slowlog.threshold.index.warn": "10s",
        "index.merge.policy.max_merged_segment": "2gb",
        "index.merge.policy.segments_per_tier": "24",
        "index.number_of_replicas": "1",
        "index.number_of_shards": "3",
        "index.optimize_auto_generated_id": "true",
        "index.refresh_interval": "10s",
        "index.routing.allocation.total_shards_per_node": "-1",
        "index.search.slowlog.threshold.fetch.debug": "500ms",
        "index.search.slowlog.threshold.fetch.info": "800ms",
        "index.search.slowlog.threshold.fetch.trace": "200ms",
        "index.search.slowlog.threshold.fetch.warn": "1s",
        "index.search.slowlog.threshold.query.debug": "2s",
        "index.search.slowlog.threshold.query.info": "5s",
        "index.search.slowlog.threshold.query.trace": "500ms",
        "index.search.slowlog.threshold.query.warn": "10s",
        // "index.translog.durability": "async",
        // "index.translog.flush_threshold_size": "5000mb",
        // "index.translog.sync_interval": "30m",
        "index.unassigned.node_left.delayed_timeout": "7200m",
        // "index.mapping.total_fields.limit": 1000000,
        "index.translog.durability": "request",

		// "number_of_shards": 1,
		// "number_of_replicas": 0
	},
	"mappings":{
		"doc":{
			"properties": {
                "doc_c_time": {
                    "type": "keyword"
                },
                "doc_i_time": {
                    "type": "keyword"
                },
                "file": {
                    "type": "text",
                    "analyzer": "test_analyzer",
                    "search_analyzer": "test_analyzer"
                },
                "file_term": {
                    "type": "keyword"
                },
                "is_file": {
                    "type": "boolean"
                },
                "is_folder": {
                    "type": "boolean"
                },
                "path": {
                    "type": "keyword"
                },
                "size": {
                    "type": "long"
                },
                "uid": {
                    "type": "keyword"
                },
                "gid": {
                    "type": "keyword"
                },
                "ctime": {
                    "type": "date",
                    "format": "epoch_second"
                },
                "mtime": {
                    "type": "date",
                    "format": "epoch_second"
                },
                "atime": {
                    "type": "date",
                    "format": "epoch_second"
                },
                'snapshot_id': {
                    "type": "keyword"
                },
                "cc_id": {
                    "type": "keyword"
                },
                "cc_name": {
                    "type": "keyword"
                },
                "tenant": {
                    "type": "keyword"
                },
                "last_used_time": {
                    "type": "date",
                    "format": "epoch_second"
                },
                "app_type": {
                    "type": "keyword"
                },
                "denied": {
                    "type": "keyword"
                },
                "app_id": {
                    "type": "keyword"
                },
                "app_name": {
                    "type": "keyword"
                },
                "file_id": {
                    "type": "keyword"
                },
                "allowed": {
                    "type": "keyword"
                }
            }
		}
	}
}`

func randomDoc() map[string]interface{} {
	doc := map[string]interface{}{
		"doc_c_time":     utils.GetCurrentTimeUnix(),
		"cc_id":          utils.GetUUID(),
		"cc_name":        utils.GetRandomString(15),
		"tenant":         utils.GetRandomString(5),
		"name":           fmt.Sprintf("%s.%s", utils.GetRandomString(10), utils.GetRandomString(3)),
		"name_term":      fmt.Sprintf("%s.%s", utils.GetRandomString(10), utils.GetRandomString(3)),
		"is_file":        true,
		"path":           []string{"", "/", "/dir", fmt.Sprintf("/dir%d", utils.GetRandomInt(1, 100))}[utils.GetRandomInt(0, 4)],
		"last_used_time": utils.GetCurrentTimeUnix(),
		"file_system":    utils.GetRandomString(5),
		"atime":          utils.GetCurrentTimeUnix(),
		"mtime":          utils.GetCurrentTimeUnix(),
		"ctime":          utils.GetCurrentTimeUnix(),
		"size":           utils.GetRandomInt(1, 1000000),
		"is_folder":      false,
		"app_type":       "test Index & Search",
		"uid":            utils.GetRandomInt(0, 10),
		"denied":         []string{},
		"app_id":         utils.GetUUID(),
		"app_name":       utils.GetRandomString(10),
		"gid":            utils.GetRandomInt(0, 10),
		"doc_i_time":     utils.GetCurrentTimeUnix(),
		"file_id":        utils.GetRandomDigit(32),
		"file":           utils.GetRandomString(20),
		"allowed":        []string{"FULL"},
	}

	return doc
}
