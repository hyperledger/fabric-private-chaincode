/*
   Copyright IBM Corp. All Rights Reserved.

   SPDX-License-Identifier: Apache-2.0
*/

package storage

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"

	"github.com/go-redis/redis"
)

const DefaultRedisPort = 6379

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

type Client struct {
	redis *redis.Client
}

func NewClient() *Client {
	host := getEnv("REDIS_HOST", "localhost")
	port := getEnv("REDIS_PORT", strconv.Itoa(DefaultRedisPort))
	password := getEnv("REDIS_PASSWORD", "")

	return &Client{redis: redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password, // no password set
		DB:       0,        // use default DB
	})}
}

func (c *Client) Upload(data []byte) (string, error) {
	hashedContent := sha256.Sum256(data)
	encodedContent := base64.StdEncoding.EncodeToString(data)
	key := base64.StdEncoding.EncodeToString(hashedContent[:])

	if err := c.redis.Set(key, encodedContent, 0).Err(); err != nil {
		return "", err
	}

	fmt.Printf("PatientData successfully uploaded to storage service!\nkey: %s\nvalue: %s\n", key, encodedContent)

	return key, nil
}
