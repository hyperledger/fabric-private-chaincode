/*
   Copyright IBM Corp. All Rights Reserved.
   Copyright 2021 Intel Corporation

   SPDX-License-Identifier: Apache-2.0
*/

package storage

import (
	"context"
	"os"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func newRedisClient() *redis.Client {
	var (
		host     = getEnv("REDIS_HOST", "localhost")
		port     = getEnv("REDIS_PORT", "6379")
		password = getEnv("REDIS_PASSWORD", "")
	)

	return redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password, // no password set
		DB:       0,        // use default DB
	})
}

func Get(key []byte) (value []byte, e error) {
	rdb := newRedisClient()

	val, err := rdb.Get(ctx, string(key)).Bytes()
	if err != nil {
		return nil, err
	}

	return val, nil
}

func Set(key string, value string) (e error) {
	rdb := newRedisClient()

	err := rdb.Set(ctx, key, value, 0).Err()
	if err != nil {
		return err
	}
	return nil
}
