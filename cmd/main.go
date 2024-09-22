package main

import (
	"fmt"
	"log"
	"os"
	"time"

	distributedstore "github.com/SangharshSeth/distributed-kv-store"
	"github.com/SangharshSeth/distributed-kv-store/pkg/stastistics"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
	logger.SetPrefix("\033[32m[INFO]\033[0m ")
}

const (
	Reset = "\033[0m"
	Blue  = "\033[34m"
	Green = "\033[32m"
)

func main() {
	partitionSize := 3
	var connectionsWithUptime = make(map[string]time.Time)

	newStatisticsStore := stastistics.NewStatisticsStore(
		connectionsWithUptime,
	)

	fmt.Println(Green + "Welcome to Distributed Key-Value Store!" + Reset)

	newStore := distributedstore.NewDistributedKVStore(
		"0.0.0.0:9090",
		newStatisticsStore,
		partitionSize,
	)

	newStore.LoadDataFromAOFFile()
	newStore.StartSystem()
}
