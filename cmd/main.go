package main

import (
	"distributed-store/internal/statistics"
	"distributed-store/internal/store"
	"time"
)

func main() {

	storageEngine := make(map[string]string)
	var connectionsWithUptime = make(map[string]time.Time)

	newStatisticsStore := statistics.NewStatisticsStore(
		connectionsWithUptime,
	)
	newStore := store.NewDistributedKVStore(
		storageEngine,
		"localhost:9090",
		newStatisticsStore,
	)

	newStore.StartSystem()
}
