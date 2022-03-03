package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-redis/redis/v8"
)

type Redis struct {
	client *redis.Client
}

func NewRedis(host, password string) *Redis {
	rdb := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password,
		DB:       0,
	})

	return &Redis{client: rdb}
}

func (r *Redis) Bootstrap(key string) error {
	_, err := r.client.Get(context.Background(), key).Result()
	if err != nil {
		if err == redis.Nil {
			return r.client.Set(context.Background(), key, "", 0).Err()
		}
	}

	return nil
}

func (r *Redis) Get(key string) (string, error) {
	val, err := r.client.Get(context.Background(), key).Result()
	return val, err
}

func (r *Redis) Set(key, value string) error {
	if err := r.client.Set(context.Background(), key, value, 0).Err(); err != nil {
		return err
	}
	return nil
}

func (r *Redis) IsValidKey(key string) error {
	if key != strings.ToLower(key) {
		return fmt.Errorf("key %q must be only alphabetic characters, all lowercase", key)
	}

	if !alphaOnly(key) {
		return fmt.Errorf("key %q must be only alphabetic characters, all lowercase", key)
	}

	return nil
}

func (r *Redis) ConfigString() string {
	return fmt.Sprintf("Connecting to redis server: %q", r.client.Options().Addr)
}
