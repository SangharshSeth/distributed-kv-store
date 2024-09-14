package store

import (
	"bufio"
	"fmt"
	"github.com/SangharshSeth/distributed-kv-store/internal/statistics"
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
	//Opening the file in Append/Write Mode makes the cursor go to end of the file
	//So to read a file you have to create a new pointer in read mode or seek to the beginning of file

	AOFLogFileName, err := os.OpenFile("AOF.txt", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
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
	_, err := d.AOFLogFileName.Seek(0, 0)
	if err != nil {
		slog.Error("Failed to seek to the beginning of file")
	}
	scanner := bufio.NewScanner(d.AOFLogFileName)

	for scanner.Scan() {
		command := scanner.Text()
		d.ProcessCommand(command, true)
	}
}

func (d *DistributedKVStore) StartSystem() {
	tcpListener, err := net.Listen("tcp", d.tcpServerAddress)
	defer tcpListener.Close()

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	slog.Info("TCP Server listening on ", slog.String("address", d.tcpServerAddress))

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
		value, exists := d.Get(inputSlice[1])
		if exists {
			return fmt.Sprintf("%s\n", value)
		}
		return "NOT FOUND\n"
	case "DEL":
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
	return int(hash) % d.PartitionSize
}
func (d *DistributedKVStore) ViewPartitionWiseData() {
	var totalElements int
	partitionSizes := make(map[int]int, len(d.dataStore))

	for index, partition := range d.dataStore {
		d.mutexLock[index].RLock()
		size := len(partition)
		partitionSizes[index] = size
		totalElements += size
		d.mutexLock[index].RUnlock()
	}

	slog.Info("Partition-wise data summary",
		slog.Int("total_partitions", len(d.dataStore)),
		slog.Int("total_elements", totalElements),
		slog.Any("partition_sizes", partitionSizes))
}

func (d *DistributedKVStore) ShutDown() {
	fmt.Println("Entered Shutting Down the Server")

	d.waitGroup.Wait()
	d.analyticsStorage.DisplayStatsInTerminal()
	//d.ViewPartitionWiseData()
	fmt.Println("Shutting Down....")
	os.Exit(0)
}
