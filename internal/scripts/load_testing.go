package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"runtime"
	"sync"
	"time"
)

func main() {
	address := "localhost:9090"
	numConnections := 100 // Number of concurrent connections
	var wg sync.WaitGroup

	for i := 0; i < numConnections; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			conn, err := net.DialTimeout("tcp", address, 5*time.Second)
			if err != nil {
				log.Printf("Failed to connect: %v", err)
				return
			}
			defer conn.Close()

			// Set a deadline for the entire connection
			conn.SetDeadline(time.Now().Add(10 * time.Second))

			// Example command
			fmt.Fprintf(conn, "SET key%d value%d\n", index, index)

			// Read response
			scanner := bufio.NewScanner(conn)
			for scanner.Scan() {
				fmt.Printf("Connection %d received: %s\n", index, scanner.Text())
				// Reset the deadline for each successful read
				conn.SetDeadline(time.Now().Add(10 * time.Second))
			}

			if err := scanner.Err(); err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					fmt.Printf("Connection %d timed out\n", index)
				} else {
					fmt.Printf("Connection %d error: %v\n", index, err)
				}
			}
		}(i)
	}

	// Wait for all connections to finish
	wg.Wait()

	// Print the number of goroutines
	fmt.Printf("Number of goroutines: %d\n", runtime.NumGoroutine())
}
