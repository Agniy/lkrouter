package redisdb

import (
	"fmt"
	"github.com/go-redis/redis"
	"lkrouter/config"
	"log"
	"os"
	"sync"
	"time"
)

var clientInstance *redis.Client
var clientInstanceError error
var redisOnce sync.Once

// Get redis client
func GetRedisClient() (*redis.Client, error) {
	logger := log.New(os.Stdout, "GetRedisClient: ", 0)
	redisOnce.Do(func() {
		cfg := config.GetConfig()
		client := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", cfg.RedisConfig.RedisHost, cfg.RedisConfig.RedisPort),
			Password: "",
			DB:       0,
			//TLSConfig: &tls.Config{
			//	MinVersion: tls.VersionTLS12,
			//	//Certificates: []tls.Certificate{cert}
			//},
		})
		_, err := client.Ping().Result()
		if err != nil {
			logger.Printf("client.Ping().Result() err: %v", err)
			clientInstanceError = err
		}
		clientInstance = client
	})
	return clientInstance, clientInstanceError
}

type RedisClient struct {
	*redis.Client
}

func NewRedisClient() (*RedisClient, error) {
	logger := log.New(os.Stdout, "NewRedisClient: ", 0)
	client, err := GetRedisClient()
	if err != nil {
		logger.Printf("NewRedisClient err:", err)
		return nil, err
	}
	return &RedisClient{client}, nil
}

func (r *RedisClient) Set(key string, value interface{}, timeout time.Duration) error {
	logger := log.New(os.Stdout, "RedisClient.Set: ", 0)
	err := r.Client.Set(key, value, timeout).Err()
	if err != nil {
		logger.Printf("r.Client.Set err: %v", err)
		return err
	}

	return nil
}

func (r *RedisClient) Get(key string) (interface{}, error) {
	logger := log.New(os.Stdout, "RedisClient.Get: ", 0)

	if r.Client == nil {
		return nil, fmt.Errorf("r.Client is nil")
	}

	val, err := r.Client.Get(key).Result()
	if err != nil {
		logger.Printf("r.Client.Get err: %v", err)
		return nil, err
	}
	return val, nil
}

func (r *RedisClient) Del(key string) error {
	err := r.Client.Del(key).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisClient) HSet(key string, field string, value interface{}, timeout time.Duration) error {
	logger := log.New(os.Stdout, "RedisClient.HSet: ", 0)
	err := r.Client.HSet(key, field, value).Err()
	if err != nil {
		logger.Printf("r.Client.HSet err: %v", err)
		return err
	}
	err = r.Client.Expire(key, timeout).Err()
	if err != nil {
		logger.Printf("r.Client.Expire err: %v", err)
		return err
	}

	return nil
}

func (r *RedisClient) HGet(key string, field string) (interface{}, error) {
	logger := log.New(os.Stdout, "RedisClient.HGet: ", 0)
	val, err := r.Client.HGet(key, field).Result()
	if err != nil {
		logger.Printf("r.Client.HGet err: %v", err)
		return nil, err
	}
	return val, nil
}

func (r *RedisClient) HDel(key string, field string) error {
	err := r.Client.HDel(key, field).Err()
	if err != nil {
		return err
	}
	return nil
}

func Set(key string, value interface{}, timeout time.Duration) error {
	logger := log.New(os.Stdout, "Set: ", 0)
	client, err := NewRedisClient()
	if err != nil {
		logger.Printf("Set err, %v", err)
		return err
	}
	return client.Set(key, value, timeout)
}

func Get(key string) (interface{}, error) {
	logger := log.New(os.Stdout, "Get: ", 0)
	client, err := NewRedisClient()
	if err != nil {
		logger.Printf("NewRedisClient err, %v", err)
		return nil, err
	}
	return client.Get(key)
}

func Del(key string) error {
	logger := log.New(os.Stdout, "Del: ", 0)
	client, err := NewRedisClient()
	if err != nil {
		logger.Printf("Del err, %v", err)
		return err
	}
	return client.Del(key)
}

func HSet(key string, field string, value interface{}, timeout time.Duration) error {
	logger := log.New(os.Stdout, "HGet: ", 0)
	client, err := NewRedisClient()
	if err != nil {
		logger.Printf("HSet err, %v", err)
		return err
	}
	return client.HSet(key, field, value, timeout)
}

func HGet(key string, field string) (interface{}, error) {
	logger := log.New(os.Stdout, "HGet: ", 0)
	client, err := NewRedisClient()
	if err != nil {
		logger.Printf("NewRedisClient err, %v", err)
		return nil, err
	}
	return client.HGet(key, field)
}
