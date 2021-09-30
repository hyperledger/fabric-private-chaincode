package container

import (
	"fmt"
	"testing"
	"time"

	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/storage"
	"github.com/stretchr/testify/assert"
)

func TestRedis(t *testing.T) {
	redis := &Container{
		Image:    "redis",
		CMD:      nil,
		Name:     "redis-container",
		HostIP:   "localhost",
		HostPort: "6379",
	}

	err := redis.Start()
	assert.NoError(t, err)

	fmt.Println("sleep")
	time.Sleep(20 * time.Second)
	fmt.Println("continue")

	c := storage.NewClient()
	handle, err := c.Upload([]byte("some data"))
	assert.NoError(t, err)
	assert.NotEmpty(t, handle)

	err = redis.Stop()
	assert.NoError(t, err)
}
