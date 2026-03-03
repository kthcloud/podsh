package syncdb

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/go-redis/redis"
	"github.com/kthcloud/podsh/internal/cache"
	"github.com/kthcloud/podsh/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type periodicSyncer struct {
	interval time.Duration

	mongoColl *mongo.Collection
	redis     *redis.Client

	lastSync time.Time
}

func newPeriodicSyncer(interval time.Duration, coll *mongo.Collection, redis *redis.Client) *periodicSyncer {
	ps := &periodicSyncer{
		interval:  interval,
		mongoColl: coll,
		redis:     redis,
	}

	return ps
}

func (w *periodicSyncer) Start(ctx context.Context) error {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := w.syncOnce(ctx); err != nil {
				return err
			}
		}
	}
}

func (w *periodicSyncer) syncOnce(ctx context.Context) error {
	cursor, err := w.mongoColl.Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var user struct {
			ID       string `bson:"id"`
			Username string `bson:"username"`
			//	Role       string `bson:"effectiveRole.name"`
			Admin      bool `bson:"isAdmin"`
			PublicKeys []struct {
				Key []byte `bson:"key"`
			} `bson:"publicKeys"`
		}

		if err := cursor.Decode(&user); err != nil {
			log.Println("err decode", err)
			continue
		}

		identity := models.Identity{
			UserID:   user.ID,
			Username: user.Username,
			//	Role:     user.Role,
			Admin: user.Admin,
		}

		data, err := json.Marshal(identity)
		if err != nil {
			log.Println("err marshal", err)
			continue
		}

		for _, pk := range user.PublicKeys {
			norm, err := cache.NormalizePublicKey(pk.Key)
			if err != nil {
				log.Println("Failed to normalize public key, skipping")
				continue
			}
			key := cache.ComputeKey(norm)
			ttl := w.interval + 5*time.Second
			if err := w.redis.Set(key, data, ttl).Err(); err != nil {
				log.Println("Redis set error:", err)
			}
		}
	}

	w.lastSync = time.Now()
	return nil
}
