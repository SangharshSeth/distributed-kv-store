package main

import (
	"distributed-store/internal/statistics"
	"distributed-store/internal/store"
	"time"
)

func main() {

	partitionSize := 3
	var connectionsWithUptime = make(map[string]time.Time)

	newStatisticsStore := statistics.NewStatisticsStore(
		connectionsWithUptime,
	)
	newStore := store.NewDistributedKVStore(
		"localhost:9090",
		newStatisticsStore,
		partitionSize,
	)

	newStore.StartSystem()
}
