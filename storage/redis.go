package storage

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/go-redis/redis/v8"
)

type Redis struct {
	client *redis.Client
}

func NewRedis(host, password, serverName string, hasTLS bool) (*Redis, error) {
	opts := &redis.Options{
		Addr:     host,
		Password: password,
		DB:       0,
	}

	if serverName != "" && !hasTLS {
		return nil, fmt.Errorf("server name has been set, however TLS connectivity is disabled")
	}

	if hasTLS {
		if serverName != "" {
			opts.TLSConfig = &tls.Config{
				MinVersion: tls.VersionTLS12,
				ServerName: serverName,
			}
		} else {
			opts.TLSConfig = &tls.Config{
				MinVersion: tls.VersionTLS12,
			}
		}
	}

	return &Redis{client: redis.NewClient(opts)}, nil
}

func (r *Redis) Bootstrap(key string) error {
	_, err := r.client.Get(context.Background(), key).Result()
	if err != nil {
		if err == redis.Nil {
			return r.client.Set(context.Background(), key, "", 0).Err()
		}

		return err
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
	return fmt.Sprintf("Redis server: %q", r.client.Options().Addr)
}
