package redis

import (
	"context"
	"errors"
	"github.com/baobao233/gorder/common/logging"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"time"
)

func SetNX(ctx context.Context, client *redis.Client, key, value string, ttl time.Duration) (err error) {
	now := time.Now()
	defer func() {
		l := logrus.WithContext(ctx).WithFields(logrus.Fields{
			"now":         now,
			"key":         key,
			"value":       value,
			logging.Error: err,
			logging.Cost:  time.Since(now) * time.Millisecond,
		})
		if err == nil {
			l.Info("_redis_setnx_success")
		} else {
			l.Info("_redis_setnx_error")
		}
	}()
	if client == nil {
		return errors.New("redis client is nil")
	}

	_, err = client.SetNX(ctx, key, value, ttl).Result()
	return err
}

func Del(ctx context.Context, client *redis.Client, key string) (err error) {
	now := time.Now()
	defer func() {
		l := logrus.WithContext(ctx).WithFields(logrus.Fields{
			"now":         now,
			"key":         key,
			logging.Error: err,
			logging.Cost:  time.Since(now) * time.Millisecond,
		})
		if err == nil {
			l.Info("_redis_del_success")
		} else {
			l.Info("_redis_del_error")
		}
	}()
	if client == nil {
		return errors.New("redis client is nil")
	}

	_, err = client.Del(ctx, key).Result()
	return err
}
