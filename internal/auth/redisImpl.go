package auth

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/go-redis/redis"
	"github.com/kthcloud/podsh/internal/cache"
	"github.com/kthcloud/podsh/internal/sshd"
)

type RedisPublicKeyAuthenticator struct {
	redisClient *redis.Client
}

func NewRedisPublicKeyAuthenticator(client *redis.Client) *RedisPublicKeyAuthenticator {
	return &RedisPublicKeyAuthenticator{
		redisClient: client,
	}
}

func (r *RedisPublicKeyAuthenticator) Authenticate(ctx context.Context, meta sshd.ConnMetadata, pubKey []byte) (*sshd.Identity, error) {
	key := cache.ComputeKey(pubKey)

	client := r.redisClient.WithContext(ctx)

	val, err := client.Get(key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("identity not found")
		}
		return nil, err
	}

	var identity sshd.Identity
	if err := json.Unmarshal([]byte(val), &identity); err != nil {
		return nil, err
	}

	return &identity, nil
}
