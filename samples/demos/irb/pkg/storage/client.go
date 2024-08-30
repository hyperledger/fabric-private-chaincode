/*
   Copyright IBM Corp. All Rights Reserved.

   SPDX-License-Identifier: Apache-2.0
*/

package storage

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/go-redis/redis"
)

const DefaultRedisPort = 6379

type Client struct {
	redis *redis.Client
}

type config struct {
	host     string
	port     int
	password string
}

func WithHost(host string) func(*config) {
	return func(c *config) {
		c.host = host
	}
}

func WithPort(port int) func(*config) {
	return func(c *config) {
		c.port = port
	}
}

func WithPassword(password string) func(*config) {
	return func(c *config) {
		c.password = password
	}
}

func NewClient(options ...func(*config)) *Client {
	c := &config{}
	for _, apply := range options {
		apply(c)
	}

	return &Client{redis: redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", c.host, c.port),
		Password: c.password, // no password set
		DB:       0,          // use default DB
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
