package syncdb

import (
	"time"

	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/mongo"
)

type KVSyncWorker struct {
	mongoColl *mongo.Collection
	redis     *redis.Client
	syncer
}

// NewKVSyncWorker auto-selects Change Stream or periodic mode
func NewKVSyncWorker(mongoColl *mongo.Collection, redis *redis.Client, interval time.Duration) *KVSyncWorker {
	worker := &KVSyncWorker{
		mongoColl: mongoColl,
		redis:     redis,
	}

	//if mongoColl.Database().Client().Options().ReplicaSet != nil {
	//	worker.syncer = newWatchSyncer()
	//} else {
	worker.syncer = newPeriodicSyncer(interval, mongoColl, redis)
	//}

	return worker
}
