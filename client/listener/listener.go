package listener

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"megalink/gateway/client/handler"
	"megalink/gateway/client/types"
	"time"
)

const (
	// ReadBufferSize max 2 bytes message size.
	readBufferSize = 1024
	// TimeoutSeconds max seconds while reading message.
	timeoutSeconds = 30
	// ReadTimeout max duration while reading message.
	readTimeout = time.Duration(timeoutSeconds) * time.Second
	// Tags.
)

// Listener listens to a connection and reads for new bytes coming through.
type Listener struct {
	// Conn connection to start listening from.
	Conn io.ReadWriter
	// ReadBuffer max buffer size to read from connection
	ReadBuffer int
	// Handler of ISO8583 message
	Handler handler.MessageHandlerFunc
	// ReadTimeout read timeout to receive a complete ISO8583 message
	ReadTimeout time.Duration
	// ErrHandler handles critical errors.
	ErrHandler handler.ErrorHandler
	// GtwDynamoConfig handles dynamoConfigGtw.
	EnvVars *types.EnvVars
}

// NewListener creates a new listener with some defaults.
func NewListener(
	conn io.ReadWriter,
	handler handler.MessageHandlerFunc,
	errorHandler handler.ErrorHandler,
	envVars *types.EnvVars) *Listener {
	return &Listener{
		Conn:        conn,
		Handler:     handler,
		ReadBuffer:  readBufferSize,
		ReadTimeout: readTimeout,
		ErrHandler:  errorHandler,
		EnvVars:     envVars,
	}
}

// Listen listens to a connection for new arriving bytes with a timeout.
func (ls *Listener) Listen(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("\nListener | Listen", "Shutting down listener")
			return
		default:
			readCtx, cancel := context.WithTimeout(ctx, ls.ReadTimeout)
			defer cancel()

			header := make([]byte, 4) // Assume the message length is encoded in the first 4 bytes
			if _, err := io.ReadFull(ls.Conn, header); err != nil {
				fmt.Println("\nListener | Failed to read message header:", err)
				//they have sleep here
				time.Sleep(time.Second)
				continue
			}

			messageLength := binary.BigEndian.Uint32(header)

			bytesToRead := int(messageLength)
			tmpData := make([]byte, readBufferSize)

			bufferData := bytes.NewBuffer(make([]byte, 0, messageLength))

			done := make(chan error, 1)

			go func() {
			readLoop:
				for bytesToRead > 0 {
					select {
					case <-readCtx.Done():
						done <- readCtx.Err()
						break readLoop
					default:
						nBytes, err := io.ReadFull(ls.Conn, tmpData[:bytesToRead])
						if err == io.EOF || err == io.ErrUnexpectedEOF {
							continue
						}
						if err != nil {
							done <- err
							break readLoop
						}
						bytesToRead -= nBytes
						_, _ = bufferData.Write(tmpData[:nBytes])
					}
				}

				// Decode server response from JSON
				var serverResponse types.ServerResponse
				if err := json.Unmarshal(bufferData.Bytes(), &serverResponse); err != nil {
					done <- fmt.Errorf("failed to unmarshal server response: %w", err)
					return
				}

				fmt.Println("\nReceived server response:", serverResponse.ServerResponse)
				done <- ls.Handler(ls.Conn, &serverResponse)
			}()

			select {
			case <-readCtx.Done():
				fmt.Println("Listener | Read timeout")
				continue
			case err := <-done:
				if err != nil {
					fmt.Println("Listener | Handler error:", err)
					continue
				}
			}
		}
	}
}
