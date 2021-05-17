/*
   Copyright IBM Corp. All Rights Reserved.
   Copyright 2021 Intel Corporation

   SPDX-License-Identifier: Apache-2.0
*/

package storage

import (
	"context"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

func newRedisClient() *redis.Client {
	var (
		host     = "localhost"
		port     = "6379"
		password = ""
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

func Set(key []byte, value []byte) (e error) {
	rdb := newRedisClient()

	err := rdb.Set(ctx, string(key), string(value), 0).Err()
	if err != nil {
		return err
	}
	return nil
}
