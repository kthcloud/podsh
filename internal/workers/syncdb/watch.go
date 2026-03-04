package syncdb

import (
	"context"
	"encoding/json"
	"log"

	"github.com/go-redis/redis"
	"github.com/kthcloud/podsh/internal/cache"
	"github.com/kthcloud/podsh/internal/models"
	"go.mongodb.org/mongo-driver/mongo"
)

type watchSyncer struct {
	mongoColl *mongo.Collection
	redis     *redis.Client
}

func (w *watchSyncer) Start(ctx context.Context) error {
	pipeline := mongo.Pipeline{}
	cs, err := w.mongoColl.Watch(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cs.Close(ctx)

	for cs.Next(ctx) {
		var event map[string]interface{}
		if err := cs.Decode(&event); err != nil {
			// maybe return err here
			continue
		}
		w.handleChangeEvent(ctx, event)
	}

	return ctx.Err()
}

// handleChangeEvent parses the event and upserts into Redis
func (w *watchSyncer) handleChangeEvent(ctx context.Context, event map[string]interface{}) {
	fullDoc, ok := event["fullDocument"].(map[string]interface{})
	if !ok {
		return
	}

	userID, _ := fullDoc["id"].(string)
	username, _ := fullDoc["username"].(string)
	// role, _ := fullDoc["effecitveRole"].(string)
	admin, _ := fullDoc["isAdmin"].(bool)
	pubKeys, _ := fullDoc["publicKeys"].([]interface{})

	identity := models.Identity{
		UserID:   userID,
		Username: username,
		//	Role:     role,
		Admin: admin,
	}

	data, err := json.Marshal(identity)
	if err != nil {
		return
	}

	for _, pk := range pubKeys {
		pkMap, ok := pk.(map[string]interface{})
		if !ok {
			continue
		}
		keyStr, ok := pkMap["key"].(string)
		if !ok {
			continue
		}
		norm, err := cache.NormalizePublicKey([]byte(keyStr))
		if err != nil {
			log.Println("error when normalizing pubkey, skipping", err)
			continue
		}
		key := cache.ComputeKey(norm)
		if err := w.redis.Set(key, data, 0).Err(); err != nil {
			log.Println("Redis set error:", err)
		}
	}
}
