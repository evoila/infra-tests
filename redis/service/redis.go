package service

import (
	"errors"
	"github.com/briandowns/spinner"
	"github.com/evoila/infraTESTure/config"
	"github.com/go-redis/redis"
	"log"
	"math/rand"
	"strconv"
	"time"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var isCluster bool
var goRedisSingleNodeClient *redis.Client
var goRedisClusterClient *redis.ClusterClient
var spin *spinner.Spinner

type RedisConnectionConfig struct {
	Addresses []string
	Password string
	DB int
}

// Create a redis client for a single node redis deployment
func newRedisSingleNodeClient(config *RedisConnectionConfig) *redis.Client {
	var goRedisSingleNodeClient = redis.NewClient(&redis.Options{
		Addr:     config.Addresses[0],
		Password: config.Password,
		DB:       config.DB,
	})

	_, err := goRedisSingleNodeClient.Ping().Result()
	if err != nil {
		panic(err)
	}

	return goRedisSingleNodeClient
}

// Create a redis client for a redis cluster deployment
func newRedisClusterClient(config *RedisConnectionConfig) *redis.ClusterClient {
	var goRedisClusterClient = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    config.Addresses,
		Password: config.Password,
	})

	_, err := goRedisClusterClient.Ping().Result()
	if err != nil {
		panic(err)
	}

	return goRedisClusterClient
}

// Check whether redis is a single node or a cluster and call the corresponding client method
func newRedisClient(config *RedisConnectionConfig) {
	if len(config.Addresses) > 1 {
		isCluster = true
		goRedisClusterClient = newRedisClusterClient(config)
	} else {
		isCluster = false
		goRedisSingleNodeClient = newRedisSingleNodeClient(config)
	}
}

// // Check whether redis is a single node or a cluster and call the corresponding set method
func set(key string, value string, duration time.Duration) {
	if isCluster {
		goRedisClusterClient.Set(key, value, duration)
	} else {
		goRedisSingleNodeClient.Set(key, value, duration)
	}
}

// Check whether redis is a single node or a cluster and call the corresponding get method
func get(key string) string {
	if isCluster {
		return goRedisClusterClient.Get(key).Val()
	} else {
		return goRedisSingleNodeClient.Get(key).Val()
	}
}

// Check whether redis is a single node or a cluster and call the corresponding delete method
func del (key string) {
	if isCluster {
		goRedisClusterClient.Del(key)
	} else {
		goRedisSingleNodeClient.Del(key)
	}
}

// Check whether redis is a single node or a cluster and call the corresponding close method
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

func init() {
	spin = spinner.New(spinner.CharSets[33], 100*time.Millisecond)
	rand.Seed(time.Now().UnixNano())
}

type ips func() []string

// Actual service Test
func TestService(config *config.Config, ips ips) (err error) {
	log.Println("[INFO] Inserting and deleting data to redis...")

	spin.Start()

	// Get the ips & append them with the service specific port
	var addresses []string
	for _, ip := range ips() {
		addresses = append(addresses, ip + ":" + strconv.Itoa(config.Service.Port))
	}

	redisConfig := RedisConnectionConfig{
		Addresses: addresses,
		Password:  config.Service.Credentials.Password,
		DB:        0,
	}

	newRedisClient(&redisConfig)

	// Create some random key-value-pair and store them in redis
	key := randomString(rand.Intn(100))
	value := randomString(rand.Intn(100))

	set(key, value, 0)

	// Check if the key-value-pair was stored correctly
	if get(key) != value {
		spin.Stop()
		return errors.New("failed to insert data correctly")
	}

	// Delete the key-value-pair
	del(key)

	// Check if the key-value-pair was deleted correctly
	if get(key) != "" {
		spin.Stop()
		return errors.New("failed to delete data correctly")
	}

	shutdown()

	spin.Stop()
	return nil
}

// Create a random string based on the const "letters"
func randomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}
