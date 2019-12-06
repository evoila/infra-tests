package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"log"
	"time"
)

var isCluster bool
var goRedisSingleNodeClient *redis.Client
var goRedisClusterClient *redis.ClusterClient

type RedisConnectionConfig struct {
	Addresses []string
	Password string
	DB int
}

// Create a single node redis client based on a given redis config
func newRedisSingleNodeClient(config *RedisConnectionConfig) *redis.Client {
	var goRedisSingleNodeClient = redis.NewClient(&redis.Options{
		Addr:     config.Addresses[0],
		Password: config.Password,
		DB:       config.DB,
		WriteTimeout: 10*time.Second,
		ReadTimeout: 10*time.Second,
	})

	_, err := goRedisSingleNodeClient.Ping().Result()

	if err != nil {
		log.Printf("[ERROR] %v", err.Error())
	}

	return goRedisSingleNodeClient
}

// Create a cluster redis client based on a given redis config
func newRedisClusterClient(config *RedisConnectionConfig) *redis.ClusterClient {
	var goRedisClusterClient = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    config.Addresses,
		Password: config.Password,
		WriteTimeout: 10*time.Second,
		ReadTimeout: 10*time.Second,
	})

	_, err := goRedisClusterClient.Ping().Result()
	if err != nil {
		log.Printf("[ERROR] %v", err.Error())
	}

	return goRedisClusterClient
}

// Check whether the deployment has a redis single node instance or a cluster
// and call the corresponding function to create a client
func newRedisClient(config *RedisConnectionConfig) {
	if len(config.Addresses) > 1 {
		isCluster = true
		goRedisClusterClient = newRedisClusterClient(config)
	} else {
		isCluster = false
		goRedisSingleNodeClient = newRedisSingleNodeClient(config)
	}
}

func set(key string, value string, duration time.Duration) {
	if isCluster {
		 goRedisClusterClient.Set(key, value, duration)
	} else {
		goRedisSingleNodeClient.Set(key, value, duration)
	}
}

func get(key string) string {
	if isCluster {
		return goRedisClusterClient.Get(key).Val()
	} else {
		return goRedisSingleNodeClient.Get(key).Val()
	}
}

func del (key string) {
	if isCluster {
		goRedisClusterClient.Del(key)
	} else {
		goRedisSingleNodeClient.Del(key)
	}
}

func shutdown() {
	var err error
	if isCluster {
		err = goRedisClusterClient.Close()
	} else {
		err = goRedisSingleNodeClient.Close()
	}
	if err != nil {
		log.Printf("[ERROR] %v", err)
	}
}

func bulkSet(values map[string]string) {
	for key, value := range values {
		set(key, value, 0)
	}
}

func bulkSetSuccessful(values map[string]string) bool {
	for key, value := range values {
		if get(key) != value {
			return false
		}
	}
	return true
}

func deletionSuccessful(values map[string]string) bool {
	for key := range values {
		if get(key) != "" {
			return false
		}
	}
	return true
}