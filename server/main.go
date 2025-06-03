package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log"
	"math/rand"
	"megalink/gateway/shared"
	"net"
	"time"
)

func RandomZeroOrOne() int {
	rand.Seed(time.Now().UnixNano()) // Seed the random number generator with current time
	return rand.Intn(4)              // Generate a random number between 0 and 1 (inclusive)
}

func handleConnection(conn net.Conn, done chan struct{}) {
	defer conn.Close()
	fmt.Println("Handle connection")

	for {
		// Read data from connection
		data := make([]byte, 1024)
		n, err := conn.Read(data)
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Println("Client closed connection")
				break
			}
			fmt.Println("Error reading data:", err)
			return
		}

		// Decode request from JSON
		var request shared.Transaction
		err = json.Unmarshal(data[:n], &request)
		if err != nil {
			fmt.Println("Error decoding request:", err)
			return
		}
		fmt.Println("request:")
		fmt.Println(request)

		//time.Sleep(8 * time.Second)

		responses := []string{"00", "00", "00", "00"}

		// Create server response
		response := request
		response.F39 = responses[RandomZeroOrOne()]
		id, _ := uuid.NewV7()
		response.F38 = id.String()[0:6]

		// Encode response as JSON
		responseData, err := json.Marshal(response)
		if err != nil {
			fmt.Println("Error encoding response:", err)
			return
		}

		// Create a buffer to hold the length header and the response data
		responseWithHeader := make([]byte, 4+len(responseData))

		// Write the length of the response data into the first 4 bytes of the buffer
		binary.BigEndian.PutUint32(responseWithHeader[:4], uint32(len(responseData)))

		// Copy the response data into the buffer after the length header
		copy(responseWithHeader[4:], responseData)

		// Write the response with the length header back to the connection
		_, err = conn.Write(responseWithHeader)
		if err != nil {
			fmt.Println("Error writing response:", err)
			return
		}
		fmt.Println("response:")
		fmt.Println(uint32(len(responseData)))
		fmt.Println(string(responseWithHeader))
	}

	done <- struct{}{} // Signal completion through channel
}

func main() {
	// Define server address for listening on port 9090
	listenAddr := "localhost:9090"

	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatal(err) // Use log.Fatal for critical errors
	}
	defer listener.Close()

	fmt.Println("Server listening on", listenAddr)
	done := make(chan struct{})
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		fmt.Println("Connection accepted:", conn.RemoteAddr().String())
		go func() {
			handleConnection(conn, done)
		}()
		go func() {
			<-done // Wait for signal from handleConnection
		}()
	}
}
