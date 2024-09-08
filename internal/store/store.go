package store

import (
	"bufio"
	"distributed-store/internal/statistics"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

type DistributedKVStore struct {
	dataStore        map[string]string
	mutexLock        *sync.RWMutex
	tcpServerAddress string
	waitGroup        *sync.WaitGroup
	analyticsStorage *statistics.Statistics
}

func NewDistributedKVStore(dataStore map[string]string, tcpServerAddress string, statisticsStore *statistics.Statistics) *DistributedKVStore {
	return &DistributedKVStore{
		dataStore:        dataStore,
		mutexLock:        &sync.RWMutex{},
		tcpServerAddress: tcpServerAddress,
		waitGroup:        &sync.WaitGroup{},
		analyticsStorage: statisticsStore,
	}
}

func (d *DistributedKVStore) StartSystem() {
	tcpListener, err := net.Listen("tcp", d.tcpServerAddress)
	defer tcpListener.Close()
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("listening on tcp://%s", d.tcpServerAddress)

	//Handling Shutdown
	stopChannel := make(chan os.Signal, 1)
	signal.Notify(stopChannel, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-stopChannel
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		d.ShutDown()
	}()

	//Infinite Loop for keeping the Server Running
	for {
		connection, err := tcpListener.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %v", err)
			continue
		}
		fmt.Println("Accepted connection from", connection.RemoteAddr())
		d.waitGroup.Add(1)
		d.analyticsStorage.AddConnection(connection.RemoteAddr().String())
		d.analyticsStorage.DisplayStatsInTerminal()
		go d.HandleConnection(connection)
	}
}

func (d *DistributedKVStore) Set(key string, value string) string {
	fmt.Printf("Trying to write key: %s, waiting for any ongoing reads to finish...\n", key)

	d.mutexLock.Lock() // Acquire write lock
	defer d.mutexLock.Unlock()

	fmt.Printf("Writing key: %s\n", key)
	d.dataStore[key] = value
	return key
}

func (d *DistributedKVStore) Get(key string) (string, bool) {
	d.mutexLock.RLock()
	defer d.mutexLock.RUnlock()
	time.Sleep(5 * time.Second)
	value, exists := d.dataStore[key]
	return value, exists

}

func (d *DistributedKVStore) HandleConnection(connection net.Conn) {
	defer d.waitGroup.Done()
	scanner := bufio.NewScanner(connection) //bufio.NewScanner can scan NewLine

	for scanner.Scan() {
		inputLine := scanner.Text()
		response := d.ProcessCommand(inputLine)
		_, err := connection.Write([]byte(response))
		if err != nil {
			return
		}
	}
}

func (d *DistributedKVStore) ProcessCommand(inputLine string) string {
	// The server will accept the following commands:
	//
	// - SET <key> <value>: Adds or updates a key-value pair in the store.
	// - GET <key>: Retrieves the value associated with the given key.
	// - DEL <key>: Deletes the key-value pair associated with the given key.

	inputLine = strings.TrimSpace(inputLine)
	inputSlice := strings.Split(inputLine, " ")

	switch inputSlice[0] {
	case "SET":
		d.Set(inputSlice[1], inputSlice[2])
		return "OK\n"
	case "GET":
		value, exists := d.Get(inputSlice[1])
		if exists {
			return fmt.Sprintf("%s\n", value)
		}
		return "NOT FOUND\n"
	default:
		return fmt.Sprintf("unknown command: %s\n", inputSlice[0])

	}

}

func (d *DistributedKVStore) ShutDown() {
	d.waitGroup.Wait()
	fmt.Println("Shutting Down")
	os.Exit(0)
}
