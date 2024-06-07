###Golang Gateway Project
This project implements a gateway application example in Go, leveraging the Gin web framework and Zap logger for efficient logging. It consists of two main components: a server and a client. The server maintains an open TCP connection, while the client provides a web service to send messages to the server and wait for responses using channels. Additionally, the client has a listener running in a goroutine to listen for server responses and a heartbeat mechanism to send heartbeats to the server.

##Technologies Used
-Go 1.18: The programming language used for developing the application.
-Gin Web Framework: A lightweight and fast HTTP web framework for Go, used to create the web service in the client.
-Zap Logger: A high-performance, structured logging library for Go, used for logging within the client.
