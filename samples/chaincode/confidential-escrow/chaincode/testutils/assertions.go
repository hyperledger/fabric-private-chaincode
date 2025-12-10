package testutils

import (
	"encoding/json"
	"testing"

	"github.com/hyperledger-labs/cc-tools/errors"
)

func AssertNoError(t *testing.T, err errors.ICCError, args ...any) {
	t.Helper() // Go reports the failure at the caller’s line
	if err != nil {
		t.Errorf("Expected no error, got: %v %v", err, args)
	}
}

func AssertError(t *testing.T, err errors.ICCError, args ...any) {
	t.Helper()
	if err == nil {
		t.Errorf("Expected error, got nil %v", args)
	}
}

// AssertErrorStatus checks if error has expected status code
func AssertErrorStatus(t *testing.T, err errors.ICCError, expectedStatus int32, msgAndArgs ...any) {
	t.Helper()
	if err == nil {
		t.Errorf("Expected error with status %d, got nil %v", expectedStatus, msgAndArgs)
		return
	}
	if err.Status() != expectedStatus {
		t.Errorf("Expected status %d, got %d %v", expectedStatus, err.Status(), msgAndArgs)
	}
}

// AssertEqual checks if two values are equal
func AssertEqual(t *testing.T, expected, actual any, msgAndArgs ...any) {
	t.Helper()
	if expected != actual {
		t.Errorf("Expected %v, got %v %v", expected, actual, msgAndArgs)
	}
}

// AssertJSONContains checks if JSON response contains expected key-value
func AssertJSONContains(t *testing.T, jsonData []byte, key string, expectedValue any) {
	t.Helper()
	var data map[string]any
	if err := json.Unmarshal(jsonData, &data); err != nil {
		t.Errorf("Failed to unmarshal JSON: %v", err)
		return
	}

	actualValue, ok := data[key]
	if !ok {
		t.Errorf("Key '%s' not found in JSON", key)
		return
	}

	if actualValue != expectedValue {
		t.Errorf("For key '%s': expected %v, got %v", key, expectedValue, actualValue)
	}
}

// AssertStateExists checks if a key exists in mock state
func AssertStateExists(t *testing.T, mockStub *MockStub, key string) {
	t.Helper()
	if _, exists := mockStub.State[key]; !exists {
		t.Errorf("Expected key '%s' to exist in state", key)
	}
}

// AssertStateNotExists checks if a key does not exist in mock state
func AssertStateNotExists(t *testing.T, mockStub *MockStub, key string) {
	t.Helper()
	if _, exists := mockStub.State[key]; exists {
		t.Errorf("Expected key '%s' to not exist in state", key)
	}
}
