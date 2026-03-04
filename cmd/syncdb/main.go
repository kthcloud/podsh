package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
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
	redisAddr := flag.String("redis-addr", getEnv("REDIS_ADDR", "localhost:6379"), "Redis address")
	redisPassword := flag.String("redis-password", getEnv("REDIS_PASSWORD", ""), "Redis password")
	redisDB := flag.Int("redis-db", getEnvInt("REDIS_DB", 0), "Redis database number")

	mongoURI := flag.String("mongo-uri", getEnv("MONGO_URI", ""), "MongoDB URI (optional if user/password/host are set)")
	mongoUser := flag.String("mongo-user", getEnv("MONGO_USER", ""), "MongoDB username")
	mongoPassword := flag.String("mongo-password", getEnv("MONGO_PASSWORD", ""), "MongoDB password")
	mongoHost := flag.String("mongo-host", getEnv("MONGO_HOST", "localhost:27017"), "MongoDB host:port")
	mongoDB := flag.String("mongo-db", getEnv("MONGO_DB", "deploy"), "MongoDB database name")
	mongoColl := flag.String("mongo-collection", getEnv("MONGO_COLLECTION", "users"), "MongoDB collection name")

	interval := flag.Duration("interval", getEnvDuration("INTERVAL", 5*time.Second), "Sync interval")
	logLevel := flag.String("log-level", getEnv("LOG_LEVEL", "info"), "Log level (debug|info|warn|error)")
	probePort := flag.String("probe-port", getEnv("PROBE_PORT", "8080"), "Port for /healthz and /readyz")

	flag.Parse()

	finalMongoURI := *mongoURI
	if finalMongoURI == "" {
		if *mongoUser != "" && *mongoPassword != "" {
			finalMongoURI = fmt.Sprintf("mongodb://%s:%s@%s", *mongoUser, *mongoPassword, *mongoHost)
		} else {
			finalMongoURI = fmt.Sprintf("mongodb://%s", *mongoHost)
		}
	}

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

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	http.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		// Optionally, check Redis and Mongo connectivity for readiness
		if err := redisClient.Ping().Err(); err != nil {
			http.Error(w, "redis not ready", http.StatusServiceUnavailable)
			return
		}
		if err := mongoClient.Ping(ctx, readpref.Primary()); err != nil {
			http.Error(w, "mongo not ready", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	go func() {
		slog.Info("starting probe server", "port", *probePort)
		if err := http.ListenAndServe(":"+*probePort, nil); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("probe server stopped", "error", err)
			os.Exit(1)
		}
	}()

	if err := worker.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("worker stopped with error", "error", err)
		os.Exit(1)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		var val int
		fmt.Sscanf(v, "%d", &val)
		return val
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		d, err := time.ParseDuration(v)
		if err == nil {
			return d
		}
	}
	return fallback
}
