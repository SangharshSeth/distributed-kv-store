package store

import (
	"bufio"
	"distributed-store/internal/statistics"
	"fmt"
	"hash/fnv"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

type DistributedKVStore struct {
	dataStore        []map[string]string
	mutexLock        []*sync.RWMutex
	tcpServerAddress string
	waitGroup        *sync.WaitGroup
	analyticsStorage *statistics.Statistics
	partitionSize    int
}

func NewDistributedKVStore(tcpServerAddress string, statisticsStore *statistics.Statistics, partitionSize int) *DistributedKVStore {
	var dataPartitions = make([]map[string]string, partitionSize)
	var mutexLocks = make([]*sync.RWMutex, partitionSize)

	for i := 0; i < partitionSize; i++ {
		dataPartitions[i] = make(map[string]string)
		mutexLocks[i] = &sync.RWMutex{}
	}

	return &DistributedKVStore{
		dataStore:        dataPartitions,
		mutexLock:        mutexLocks,
		tcpServerAddress: tcpServerAddress,
		waitGroup:        &sync.WaitGroup{},
		analyticsStorage: statisticsStore,
		partitionSize:    partitionSize,
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
		go d.HandleConnection(connection)
	}
}

func (d *DistributedKVStore) Set(key string, value string) string {

	partitionIndex := d.HashKeyIntoPartitions(key)

	log.Printf("Setting the key in partition %d", partitionIndex)

	d.mutexLock[partitionIndex].Lock()
	d.mutexLock[partitionIndex].Unlock()

	d.dataStore[partitionIndex][key] = value
	return key
}

func (d *DistributedKVStore) Get(key string) (string, bool) {
	dataInThisPartitionIndex := d.HashKeyIntoPartitions(key)

	log.Printf("Getting the key from partiton %d", dataInThisPartitionIndex)

	d.mutexLock[dataInThisPartitionIndex].RLock()
	d.mutexLock[dataInThisPartitionIndex].RUnlock()

	value, exists := d.dataStore[dataInThisPartitionIndex][key]
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
func (d *DistributedKVStore) GetAll() {
	for key, value := range d.dataStore {
		fmt.Println(key, " ", value)
	}
	fmt.Println(len(d.dataStore))
}

func (d *DistributedKVStore) ShutDown() {
	d.waitGroup.Wait()
	d.analyticsStorage.DisplayStatsInTerminal()
	d.ViewPartitionWiseData()
	fmt.Println("Shutting Down")
	os.Exit(0)
}

func (d *DistributedKVStore) HashKeyIntoPartitions(key string) int {
	hasher := fnv.New32()
	hasher.Write([]byte(key))
	hash := hasher.Sum32()
	log.Printf("hash is %d, partitionSize is %d", int(hash), d.partitionSize)
	return int(hash) % d.partitionSize
}

func (d *DistributedKVStore) ViewPartitionWiseData() {
	for index, value := range d.dataStore {
		log.Printf("Current partition %d has %d elements\n", index, len(value))
	}
}
