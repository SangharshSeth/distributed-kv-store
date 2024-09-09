package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"net"
	"time"
)

func main() {
	address := "localhost:9090"
	numConnections := 150 // Number of sequential connections

	startTime := time.Now()

	for i := 0; i < numConnections; i++ {
		// Measure the time per connection
		connStart := time.Now()

		conn, err := net.DialTimeout("tcp", address, 2*time.Second) // Reduced timeout
		if err != nil {
			log.Printf("Failed to connect: %v", err)
			return
		}

		key := randomString(8)
		value := randomString(12)

		// Send SET command using conn.Write directly for better performance
		_, err = conn.Write([]byte(fmt.Sprintf("SET %s %s\n", key, value)))
		if err != nil {
			log.Printf("Failed to send data: %v", err)
			conn.Close()
			continue
		}

		// Read the response, reducing buffer size for speed improvement
		response := make([]byte, 512) // Small buffer for simple commands
		_, err = conn.Read(response)
		if err != nil {
			log.Printf("Failed to read response: %v", err)
		} else {
			fmt.Printf("Connection %d received: %s\n", i, string(response))
		}

		conn.Close()

		// Print time taken for the current connection
		fmt.Printf("Connection %d took %s\n", i, time.Since(connStart))
	}

	// Print total time taken for the load test
	duration := time.Since(startTime)
	fmt.Printf("Load test completed in %s\n", duration)
}

// Generate a random string of the given length
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := make([]byte, n)
	_, err := rand.Read(bytes)
	if err != nil {
		log.Fatalf("Failed to generate random string: %v", err)
	}

	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	return string(bytes)
}
