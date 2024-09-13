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
func PrintBanner() {
	banner := `  ____  ______ _______ _______ ______ _____  ______ _____ _____  _____ 
 |  _ \|  ____|__   __|__   __|  ____|  __ \|  ____|  __ \_   _|/ ____|
 | |_) | |__     | |     | |  | |__  | |__) | |__  | |  | || | | (___  
 |  _ <|  __|    | |     | |  |  __| |  _  /|  __| | |  | || |  \___ \ 
 | |_) | |____   | |     | |  | |____| | \ \| |____| |__| || |_ ____) |
 |____/|______|  |_|     |_|  |______|_|  \_\______|_____/_____|_____/ VERSION 1.0 `
	fmt.Println(Blue + banner + Reset)
}

// PrintFakeProgress simulates a progress bar
func PrintFakeProgress() {
	fmt.Print("System is starting")
	for i := 0; i <= 10; i++ {
		time.Sleep(250 * time.Millisecond)
		fmt.Print(".")
	}
	fmt.Println()
}

func main() {

	partitionSize := 3
	var connectionsWithUptime = make(map[string]time.Time)

	newStatisticsStore := statistics.NewStatisticsStore(
		connectionsWithUptime,
	)
	newStore := store.NewDistributedKVStore(
		"0.0.0.0:9090",
		newStatisticsStore,
		partitionSize,
	)
	PrintBanner()

	// Simulate a progress bar
	PrintFakeProgress()

	// Welcome message
	fmt.Println(Green + "Welcome to Distributed Key-Value Store!" + Reset)
	fmt.Println("Developed by Sangharsh Seth")
	newStore.LoadDataFromAOFFile()
	newStore.StartSystem()
}
