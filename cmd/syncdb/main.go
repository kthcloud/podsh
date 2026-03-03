package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis"
	"github.com/kthcloud/podsh/internal/workers/syncdb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func main() {
	var (
		redisAddr     = flag.String("redis-addr", "localhost:6379", "Redis address")
		redisPassword = flag.String("redis-password", "", "Redis password")
		redisDB       = flag.Int("redis-db", 0, "Redis database number")

		mongoURI  = flag.String("mongo-uri", "mongodb://admin:password@localhost:27017", "MongoDB URI")
		mongoDB   = flag.String("mongo-db", "deploy", "MongoDB database name")
		mongoColl = flag.String("mongo-collection", "users", "MongoDB collection name")

		interval = flag.Duration("interval", 5*time.Second, "Sync interval")
		logLevel = flag.String("log-level", "info", "Log level (debug|info|warn|error)")
	)

	flag.Parse()

	var level slog.Level
	switch *logLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     *redisAddr,
		Password: *redisPassword,
		DB:       *redisDB,
	}).WithContext(ctx)

	if err := redisClient.Ping().Err(); err != nil {
		slog.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(*mongoURI))
	if err != nil {
		slog.Error("failed to connect to mongo", "error", err)
		os.Exit(1)
	}
	defer mongoClient.Disconnect(context.Background())

	if err := mongoClient.Ping(ctx, &readpref.ReadPref{}); err != nil {
		slog.Error("failed to ping to mongo", "error", err)
		os.Exit(1)
	}

	collection := mongoClient.
		Database(*mongoDB).
		Collection(*mongoColl)

	worker := syncdb.NewKVSyncWorker(collection, redisClient, *interval)

	slog.Info("starting syncdb worker")

	if err := worker.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("worker stopped with error", "error", err)
		os.Exit(1)
	}
}
