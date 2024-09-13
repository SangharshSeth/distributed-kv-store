package store

import (
	"bufio"
	"distributed-store/internal/statistics"
	"fmt"
	"hash/fnv"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

// DistributedKVStore is a struct representing a distributed key-value store. It allows for concurrent read and write operations.
// Maximum KEY size supported is 128 Bytes
type DistributedKVStore struct {
	dataStore        []map[string]string
	mutexLock        []*sync.RWMutex
	tcpServerAddress string
	waitGroup        *sync.WaitGroup
	analyticsStorage *statistics.Statistics
	PartitionSize    int
	AOFLogFileName   *os.File
}

func NewDistributedKVStore(tcpServerAddress string, statisticsStore *statistics.Statistics, partitionSize int) *DistributedKVStore {
	var dataPartitions = make([]map[string]string, partitionSize)
	var mutexLocks = make([]*sync.RWMutex, partitionSize)
	AOFLogFileName, err := os.Open("AOF.txt")
	if err != nil {
		slog.Error("Failed to Create or Open AOF File", slog.String("error", err.Error()))
	}

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
		PartitionSize:    partitionSize,
		AOFLogFileName:   AOFLogFileName,
	}
}

func (d *DistributedKVStore) LoadDataFromAOFFile() {
	//Read from d.AOFLogFileName
	slog.Info("Loading data from AOF File")
	scanner := bufio.NewScanner(d.AOFLogFileName)
	for scanner.Scan() {
		command := scanner.Text()
		slog.Info("Command is ", slog.String("command", command))
		d.ProcessCommand(command, true)
	}
	d.AOFLogFileName.Close()
	slog.Info("Length of map is ", slog.Int("length", len(d.dataStore)))
}

func (d *DistributedKVStore) StartSystem() {
	tcpListener, err := net.Listen("tcp", d.tcpServerAddress)
	defer tcpListener.Close()

	slog.Info("AOF File Changes are applied to the system")
	slog.Info("Storage service elements count is", slog.Int("count", len(d.dataStore)))

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("TCP Server listening on %s", d.tcpServerAddress)

	//Handling Shutdown
	stopChannel := make(chan os.Signal, 1)
	signal.Notify(stopChannel, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-stopChannel
		slog.Info("Shutting down the server........")
		d.ShutDown()
	}()
	//Infinite Loop for keeping the Server Running
	for {
		connection, err := tcpListener.Accept()
		if err != nil {
			slog.Error("failed to accept connection: %v", slog.String("error", err.Error()))
			continue
		}
		slog.Info("Accepted connection from", slog.String("client", connection.RemoteAddr().String()))
		d.waitGroup.Add(1)
		d.analyticsStorage.AddConnection(connection.RemoteAddr().String())
		go d.HandleConnection(connection)
	}
}
func (d *DistributedKVStore) Set(key string, value string) string {
	if len(key) > 128 {
		return "Key size is too large"
	}
	partitionIndex := d.HashKeyIntoPartitions(key)
	d.mutexLock[partitionIndex].Lock()
	defer d.mutexLock[partitionIndex].Unlock()
	d.dataStore[partitionIndex][key] = value
	return key
}
func (d *DistributedKVStore) Get(key string) (string, bool) {
	dataInThisPartitionIndex := d.HashKeyIntoPartitions(key)

	log.Printf("Getting the key from partiton %d", dataInThisPartitionIndex)

	d.mutexLock[dataInThisPartitionIndex].RLock()
	defer d.mutexLock[dataInThisPartitionIndex].RUnlock()

	value, exists := d.dataStore[dataInThisPartitionIndex][key]
	return value, exists

}
func (d *DistributedKVStore) Delete(key string) bool {
	partitionIndex := d.HashKeyIntoPartitions(key)
	d.mutexLock[partitionIndex].Lock()
	defer d.mutexLock[partitionIndex].Unlock()

	// Check if the key exists before deleting
	if _, exists := d.dataStore[partitionIndex][key]; exists {
		delete(d.dataStore[partitionIndex], key)
		return true
	}
	return false
}
func (d *DistributedKVStore) HandleConnection(connection net.Conn) {
	defer d.waitGroup.Done()
	scanner := bufio.NewScanner(connection) //bufio.NewScanner can scan NewLine

	for scanner.Scan() {
		inputLine := scanner.Text()
		response := d.ProcessCommand(inputLine, false)
		_, err := connection.Write([]byte(response))
		if err != nil {
			return
		}
	}
}
func (d *DistributedKVStore) ProcessCommand(inputLine string, isLoadingFromAOF bool) string {
	// The server will accept the following commands:
	//
	// - SET <key> <value>: Adds or updates a key-value pair in the store.
	// - GET <key>: Retrieves the value associated with the given key.
	// - DEL <key>: Deletes the key-value pair associated with the given key.

	inputLine = strings.TrimSpace(inputLine)
	inputSlice := strings.Split(inputLine, " ")
	if !isLoadingFromAOF && (inputSlice[0] == "SET" || inputSlice[0] == "DEL") {
		if _, err := d.AOFLogFileName.WriteString(inputLine + "\n"); err != nil {
			slog.Error("Failed to write to AOF File", slog.String("error", err.Error()))
		} else {
			d.AOFLogFileName.Sync()
			slog.Info("Wrote to AOF File", slog.String("command", inputSlice[0]))
		}
	}

	if len(inputSlice) < 2 {
		return "invalid command\n"
	}

	switch inputSlice[0] {
	case "SET":
		d.Set(inputSlice[1], inputSlice[2])
		return "OK\n"
	case "GET":
		for key, value := range d.dataStore {
			d.mutexLock[key].RLock()
			fmt.Println(key, " ", value)
		}
		value, exists := d.Get(inputSlice[1])
		if exists {
			return fmt.Sprintf("%s\n", value)
		}
		return "NOT FOUND\n"
	case "DEL":
		if !isLoadingFromAOF {
			d.AOFLogFileName.WriteString(inputLine + "\n")
		}
		deleted := d.Delete(inputSlice[1])
		if deleted {
			return "KEY DELETED\n"
		}
		return "NOT FOUND\n"
	default:
		return fmt.Sprintf("unknown command: %s\n", inputSlice[0])
	}
}
func (d *DistributedKVStore) GetAll() {
	for key, value := range d.dataStore {
		d.mutexLock[key].RLock()
		fmt.Println(key, " ", value)
	}
	fmt.Println(len(d.dataStore))
}
func (d *DistributedKVStore) HashKeyIntoPartitions(key string) int {
	//Using a non-cryptic hashing algorithm for performance
	hashEngine := fnv.New32()
	_, err := hashEngine.Write([]byte(key))
	if err != nil {
		return 0
	}
	hash := hashEngine.Sum32()
	log.Printf("hash is %d, partitionSize is %d", int(hash), d.PartitionSize)
	return int(hash) % d.PartitionSize
}
func (d *DistributedKVStore) ViewPartitionWiseData() {
	for index, value := range d.dataStore {
		log.Printf("Current partition %d has %d elements\n", index, len(value))
	}
}
func (d *DistributedKVStore) ShutDown() {
	d.waitGroup.Wait()
	d.analyticsStorage.DisplayStatsInTerminal()
	d.ViewPartitionWiseData()
	fmt.Println("Shutting Down....")
	os.Exit(0)
}
