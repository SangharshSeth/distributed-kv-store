package main

import (
	"distributed-store/internal/statistics"
	"distributed-store/internal/store"
	"fmt"
	"log"
	"os"
	"time"
)

var logger *log.Logger

func init() {
	// Initialize the global logger with colored output
	logger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

	// Example of formatting to add colored output without external package
	logger.SetPrefix("\033[32m[INFO]\033[0m ")
}

const (
	Reset = "\033[0m"
	Blue  = "\033[34m"
	Green = "\033[32m"
)

// PrintBanner prints a custom ASCII art banner with colors

// PrintFakeProgress simulates a progress bar
func main() {
	partitionSize := 3
	var connectionsWithUptime = make(map[string]time.Time)

	newStatisticsStore := statistics.NewStatisticsStore(
		connectionsWithUptime,
	)

	fmt.Println(Green + "Welcome to Distributed Key-Value Store!" + Reset)

	newStore := store.NewDistributedKVStore(
		"0.0.0.0:9090",
		newStatisticsStore,
		partitionSize,
	)

	newStore.LoadDataFromAOFFile()
	newStore.StartSystem()
}
