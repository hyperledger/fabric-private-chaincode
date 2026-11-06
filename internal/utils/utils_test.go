/*
SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testKeyCreator struct{}

func (testKeyCreator) CreateCompositeKey(objectType string, attributes []string) (string, error) {
	key := "\x00" + objectType + "\x00"
	for _, a := range attributes {
		key += a + "\x00"
	}
	return key, nil
}

func TestIsFPCCompositeKey(t *testing.T) {
	assert.False(t, IsFPCCompositeKey(""))
	assert.False(t, IsFPCCompositeKey("."))
	assert.False(t, IsFPCCompositeKey("plain"))
	assert.False(t, IsFPCCompositeKey(".asset"))
	assert.True(t, IsFPCCompositeKey(".."))
	assert.True(t, IsFPCCompositeKey(".asset.a1."))
}

func TestToFabricKey(t *testing.T) {
	for key, expected := range map[string]string{
		"":           "",
		".":          ".",
		"plain":      "plain",
		".asset.a1.": "\x00asset\x00a1\x00",
	} {
		k, err := ToFabricKey(testKeyCreator{}, key)
		assert.NoError(t, err)
		assert.Equal(t, expected, k, "key %q", key)
	}
}
