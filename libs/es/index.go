package es

import (
	"context"
	"errors"
	"math/rand"
	"putt/libs/utils"
	"sync"
	"sync/atomic"
	"time"

	"github.com/olivere/elastic/v6"
)

// Elasticsearch via BulkProcessor.

// Bulker ...
type Bulker struct {
	c           *elastic.Client
	p           *elastic.BulkProcessor
	workers     int    // BulkProcessor workers allow to be executed
	index       string // index name
	bulkSize    int    // BulkSize specifies when to flush based on the size (in bytes) of the actions currently added. default:5MB
	bulkActions int    // BulkActions specifies when to flush based on the number of actions currently added. default:1000

	beforeCalls  int64          // # of calls into before callback
	afterCalls   int64          // # of calls into after callback
	failureCalls int64          // # of successful calls into after callback
	successCalls int64          // # of successful calls into after callback
	wg           sync.WaitGroup // WaitGroup
	seq          int64          // sequential id
	stopC        chan struct{}  // stop channel for the indexer
	throttleMu   sync.Mutex     // guards the following block
	throttle     bool           // throttle (or stop) sending data into bulk processor?
}

// Run starts the Bulker.
func (b *Bulker) Run() error {
	ctx := context.Background()
	// Recreate Elasticsearch index
	if err := b.CreateIndex(CreateIndexBody); err != nil {
		return err
	}

	// Start bulk processor
	p, err := b.c.BulkProcessor().
		Workers(b.workers).              // # of workers
		BulkActions(b.bulkActions).      // # of queued requests before committed
		BulkSize(b.bulkSize).            // # of bytes in requests before committed
		FlushInterval(10 * time.Second). // autocommit every 30 seconds
		Stats(true).                     // gather statistics
		Before(b.before).                // call "before" before every commit
		After(b.after).                  // call "after" after every commit
		Do(ctx)
	if err != nil {
		return err
	}

	defer p.Close()

	b.p = p

	// Start indexer that pushes data into bulk processor
	b.stopC = make(chan struct{})
	go b.indexer()

	return nil
}

// Close the bulker.
func (b *Bulker) Close() error {
	b.stopC <- struct{}{}
	<-b.stopC
	close(b.stopC)
	return nil
}

// indexer is a goroutine that periodically pushes data into
// bulk processor unless being "throttled" or "stopped".
func (b *Bulker) indexer() {
	var stop bool

	for !stop {
		select {
		case <-b.stopC:
			stop = true
		default:
			b.throttleMu.Lock()
			throttled := b.throttle
			b.throttleMu.Unlock()

			if !throttled {
				// Sample data structure
				// doc := struct {
				// 	Seq int64 `json:"seq"`
				// }{
				// 	Seq: atomic.AddInt64(&b.seq, 1),
				// }

				logger.Info("sssss")
				doc := randomDoc()
				logger.Info("aaa")
				logger.Info(utils.Prettify(doc))
				// Add bulk request.
				// Notice that we need to set Index and Type here!
				r := elastic.NewBulkIndexRequest().Index(b.index).Type("doc").Doc(doc)
				b.p.Add(r)
			}
			// Sleep for a short time.
			time.Sleep(time.Duration(rand.Intn(7)) * time.Millisecond)
		}
	}

	b.stopC <- struct{}{} // ack stopping

	err := b.p.Start(context.Background())

	if err != nil {
		panic(err)
	}
}

// before is invoked from bulk processor before every commit.
func (b *Bulker) before(id int64, requests []elastic.BulkableRequest) {
	atomic.AddInt64(&b.beforeCalls, 1)
}

// after is invoked by bulk processor after every commit.
// The err variable indicates success or failure.
func (b *Bulker) after(id int64, requests []elastic.BulkableRequest, response *elastic.BulkResponse, err error) {
	atomic.AddInt64(&b.afterCalls, 1)

	b.throttleMu.Lock()
	if err != nil {
		atomic.AddInt64(&b.failureCalls, 1)
		b.throttle = true // bulk processor in trouble
	} else {
		atomic.AddInt64(&b.successCalls, 1)
		b.throttle = false // bulk processor ok
	}
	b.throttleMu.Unlock()
}

// Stats returns statistics from bulk processor.
func (b *Bulker) Stats() elastic.BulkProcessorStats {
	return b.p.Stats()
}

// ensureIndex creates the index in Elasticsearch.
// It will be dropped if it already exists.
func (b *Bulker) ensureIndex() error {
	ctx := context.Background()

	if b.index == "" {
		return errors.New("no index name")
	}
	exists, err := b.c.IndexExists(b.index).Do(ctx)
	if err != nil {
		return err
	}
	if exists {
		_, err = b.c.DeleteIndex(b.index).Do(ctx)
		if err != nil {
			return err
		}
	}
	_, err = b.c.CreateIndex(b.index).Do(ctx)
	if err != nil {
		return err
	}
	return nil
}

// CreateIndex ...
func (b *Bulker) CreateIndex(body string) error {
	ctx := context.Background()
	// Use the IndexExists service to check if a specified index exists.
	exists, err := b.c.IndexExists(b.index).Do(ctx)
	if err != nil {
		// Handle error
		panic(err)
	}
	if exists {
		logger.Infof("Index with name %s already exists, Skip create!", b.index)
	} else {
		// Create a new index.
		createIndex, err := b.c.CreateIndex(b.index).BodyString(body).Do(ctx)
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
