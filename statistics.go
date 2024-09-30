package distributed_store

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"os"
	"strconv"
	"sync"
	"time"
)

type Statistics struct {
	connectedClientsWithUptime map[string]time.Time
	mutex                      *sync.RWMutex // To safely access statistics concurrently
}

func NewStatisticsStore(connectedClientsWithTimeStamp map[string]time.Time) *Statistics {
	return &Statistics{
		connectedClientsWithUptime: connectedClientsWithTimeStamp,
		mutex:                      &sync.RWMutex{},
	}
}

func (s *Statistics) AddConnection(address string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.connectedClientsWithUptime[address] = time.Now()
}

func (s *Statistics) RemoveConnection(address string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.connectedClientsWithUptime, address)
}

func (s *Statistics) GetUptime(address string) (float64, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	startTime, exists := s.connectedClientsWithUptime[address]
	if !exists {
		return 0, false
	}
	uptime := time.Since(startTime).Seconds()
	return uptime, true
}

func (s *Statistics) GetConnectionData() map[string]time.Time {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.connectedClientsWithUptime
}

func (s *Statistics) DisplayStatsInTerminal() {
	// Create a new table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Serial No", "Client Address", "Connection Time", "Uptime (s)"})

	// Retrieve data and populate the table
	var serialNo int = 0
	for address, startTime := range s.GetConnectionData() {
		uptime := time.Since(startTime).Seconds()
		table.Append([]string{
			strconv.Itoa(serialNo),
			address,
			startTime.Format(time.RFC3339),
			fmt.Sprintf("%.2f", uptime),
		})
		serialNo++
	}

	table.Render()
}
