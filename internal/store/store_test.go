package store

import (
	"distributed-store/internal/statistics"
	"testing"
	"time"
)

func TestDistributedKVStore(t *testing.T) {
	// Initialize the store
	stats := statistics.NewStatisticsStore(make(map[string]time.Time))
	store := NewDistributedKVStore("localhost:8080", stats, 10)

	// Test Set and Get
	t.Run("Set and Get", func(t *testing.T) {
		key := "testKey"
		value := "testValue"

		store.Set(key, value)
		retrievedValue, exists := store.Get(key)

		if !exists {
			t.Errorf("Key %s not found after Set", key)
		}

		if retrievedValue != value {
			t.Errorf("Expected value %s, got %s", value, retrievedValue)
		}
	})

	// Test Delete
	t.Run("Delete", func(t *testing.T) {
		key := "deleteKey"
		value := "deleteValue"

		store.Set(key, value)
		deleted := store.Delete(key)

		if !deleted {
			t.Errorf("Delete operation failed for key %s", key)
		}

		_, exists := store.Get(key)
		if exists {
			t.Errorf("Key %s still exists after Delete", key)
		}
	})

	// Test large key
	t.Run("Large Key", func(t *testing.T) {
		largeKey := string(make([]byte, 129)) // 129 byte key
		result := store.Set(largeKey, "value")

		if result != "Key size is too large" {
			t.Errorf("Expected 'Key size is too large', got %s", result)
		}
	})

	// Test ProcessCommand
	t.Run("ProcessCommand", func(t *testing.T) {
		// Test SET
		setCmd := "SET testCmd testValue"
		setResult := store.ProcessCommand(setCmd)
		if setResult != "OK\n" {
			t.Errorf("Expected 'OK\\n' for SET, got %s", setResult)
		}

		// Test GET
		getCmd := "GET testCmd"
		getResult := store.ProcessCommand(getCmd)
		if getResult != "testValue\n" {
			t.Errorf("Expected 'testValue\\n' for GET, got %s", getResult)
		}

		// Test DEL
		delCmd := "DEL testCmd"
		delResult := store.ProcessCommand(delCmd)
		if delResult != "KEY DELETED\n" {
			t.Errorf("Expected 'KEY DELETED\\n' for DEL, got %s", delResult)
		}

		// Test unknown command
		unknownCmd := "UNKNOWN testCmd"
		unknownResult := store.ProcessCommand(unknownCmd)
		if unknownResult != "unknown command: UNKNOWN\n" {
			t.Errorf("Expected 'unknown command: UNKNOWN\\n', got %s", unknownResult)
		}
	})

	// Test HashKeyIntoPartitions
	t.Run("HashKeyIntoPartitions", func(t *testing.T) {
		key := "testKey"
		partition := store.HashKeyIntoPartitions(key)
		if partition < 0 || partition >= store.PartitionSize {
			t.Errorf("HashKeyIntoPartitions returned invalid partition: %d", partition)
		}
	})
}
