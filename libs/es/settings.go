package es

import (
	"fmt"
	"pzatest/libs/utils"
)

// DocType ...
const DocType = "doc"

// CreateIndexBody : for create index body, settings && mappings
// "index.translog.durability": "async",
// "index.translog.flush_threshold_size": "5000mb",
// "index.translog.sync_interval": "30m",
// "index.mapping.total_fields.limit": 1000000,
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
        "indexing.slowlog.threshold.index.debug": "2s",
        "indexing.slowlog.threshold.index.info": "5s",
        "indexing.slowlog.threshold.index.trace": "500ms",
        "indexing.slowlog.threshold.index.warn": "10s",
        "merge.policy.max_merged_segment": "2gb",
        "merge.policy.segments_per_tier": "24",
        "number_of_replicas": "1",
        "number_of_shards": "3",
        "optimize_auto_generated_id": "true",
        "refresh_interval": "10s",
        "routing.allocation.total_shards_per_node": "-1",
        "search.slowlog.threshold.fetch.debug": "500ms",
        "search.slowlog.threshold.fetch.info": "800ms",
        "search.slowlog.threshold.fetch.trace": "200ms",
        "search.slowlog.threshold.fetch.warn": "1s",
        "search.slowlog.threshold.query.debug": "2s",
        "search.slowlog.threshold.query.info": "5s",
        "search.slowlog.threshold.query.trace": "500ms",
        "search.slowlog.threshold.query.warn": "10s",
        "unassigned.node_left.delayed_timeout": "7200m",
        "translog.durability": "request",
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
	logger.Info("-12222--")
	doc := map[string]interface{}{
		"doc_c_time":     utils.GetCurrentTimeUnix(),
		"cc_id":          utils.GetUUID(),
		"cc_name":        utils.GetRandomString(15),
		"tenant":         utils.GetRandomString(5),
		"name":           fmt.Sprintf("%s.%s", utils.GetRandomString(10), utils.GetRandomString(3)),
		"name_term":      fmt.Sprintf("%s.%s", utils.GetRandomString(10), utils.GetRandomString(3)),
		"is_file":        true,
		"path":           []string{"", "/", "/dir", fmt.Sprintf("/dir%d", utils.GetRandomInt(1, 100))}[utils.GetRandomInt(0, 3)],
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
	logger.Info("---")
	logger.Info(utils.Prettify(doc))
	return doc
}
