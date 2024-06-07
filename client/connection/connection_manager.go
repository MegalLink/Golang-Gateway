// Package connection implements datafast connection logic.
package connection

import (
	"context"
	"fmt"
	"io"
	"log"
	"megalink/gateway/client/heartbeat"
	"megalink/gateway/client/sign"
	"megalink/gateway/client/types"
	"net"
	"reflect"
	"sync"
	"time"
)

const (
	connManagerTag = "ConnManager | %s"
)

type (
	// ScheduledTask represents a calendared task.
	ScheduledTask func(writer io.ReadWriter)

	// Scheduler will perform some task over the connection at some interval.
	Scheduler struct {
		Conn io.ReadWriter
	}
)

// ScheduleTask spawns a goroutine at a specified interval
// returns ticker to caller.
func (sc *Scheduler) ScheduleTask(fn ScheduledTask, interval time.Duration) *time.Ticker {
	ticker := time.NewTicker(interval)

	go func(ticker *time.Ticker) {
		for range ticker.C {
			fn(sc.Conn)
		}
	}(ticker)

	return ticker
}

type (
	// IConnManager deals with connection details with franchise.
	IConnManager interface {
		net.Conn
		SetupConnection(context.Context) error
		CloseConnection() error
		TryReconnect()
	}

	// ConnManager implements IConnManager to deal with connection to franchise.
	ConnManager struct {
		SignService       sign.ISignService
		HeartbeatService  heartbeat.IHeartbeatService
		Connection        net.Conn
		ConnectionMtx     *sync.RWMutex
		ConnectionFactory IConnFactory
		EnvVars           *types.EnvVars
	}
)

// NewConnManager initializes a new IConnManager.
func NewConnManager(
	signService sign.ISignService,
	heartbeatService heartbeat.IHeartbeatService,
	connectionFactory IConnFactory,
	envVars *types.EnvVars,
) IConnManager {
	return &ConnManager{
		SignService:       signService,
		HeartbeatService:  heartbeatService,
		Connection:        nil,
		ConnectionFactory: connectionFactory,
		ConnectionMtx:     &sync.RWMutex{},
		EnvVars:           envVars,
	}
}

// SetupConnection sets up a connection with the franchise.
func (cm *ConnManager) SetupConnection(ctx context.Context) error {
	fmt.Println("\nSetting up connection | SetupConnection ")

	conn, err := cm.ConnectionFactory.GetConnection()
	if err != nil {
		fmt.Printf("\nSetupConnection | GetConnection Error %v", err)
		return err
	}

	cm.ConnectionMtx.Lock()
	defer cm.ConnectionMtx.Unlock()

	cm.Connection = conn
	err = cm.SignService.SendSignOn(cm.Connection)
	if err != nil {
		fmt.Printf("\nSetupConnection | SendSignOn Error %v", err)
		return err
	}

	connectionMsg := fmt.Sprintf("\nConnections is UP with %s", cm.Connection.RemoteAddr().String())
	fmt.Println(connectionMsg)

	heartBeatInterval := time.Duration(cm.EnvVars.HeartSendBeatIntervalSeconds) * time.Second
	log.Println("heartBeatInterval", heartBeatInterval, cm.EnvVars.HeartSendBeatIntervalSeconds)
	go cm.setupHeartbeat(ctx, heartBeatInterval)

	return nil
}

func (cm *ConnManager) tryCloseConnection() error {
	tag := fmt.Sprintf(connManagerTag, "tryCloseConnection")
	cm.ConnectionMtx.RLock()
	defer cm.ConnectionMtx.RUnlock()

	isNil := IsNil(cm.Connection)
	fmt.Printf("\n %s | Connection is nil: %v", tag, isNil)

	if !isNil {
		if err := cm.Connection.Close(); err != nil {
			fmt.Printf("\n%s | %v", tag, err)
			return err
		}

		fmt.Println("\nConnection closed")
	}

	return nil
}

// CloseConnection tries to close current connection with franchise.
func (cm *ConnManager) CloseConnection() error {
	return cm.tryCloseConnection()
}

func (cm *ConnManager) setupHeartbeat(ctx context.Context, interval time.Duration) {
	tag := fmt.Sprintf(connManagerTag, "setupHeartbeat")
	scheduler := Scheduler{Conn: cm}

	ticker := scheduler.ScheduleTask(cm.HeartbeatService.SendEchoTest, interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case err := <-cm.HeartbeatService.GetError():
			if err != nil {
				fmt.Printf("\n%s | %v sending to reconnect", tag, err)
				go cm.TryReconnect()
				return
			}
		}
	}
}

// TryReconnect tries to establish a new connection with franchise.
// If it fails, it panics.
func (cm *ConnManager) TryReconnect() {
	tag := fmt.Sprintf(connManagerTag, "tryReconnect")

	fmt.Println("\n", tag)
	// try to gracefully close current connection if exists.
	if err := cm.tryCloseConnection(); err != nil {
		fmt.Printf("\n%s | %v close connection failed", tag, err)
	}

	// try to set up a new connection once again.
	err := cm.SetupConnection(context.Background())
	if err != nil {
		fmt.Printf("\n%s Error | %v", tag, err)
		log.Panicln(err)
	}
}

// IsNil checks if a given interface is Nil.
func IsNil(i interface{}) bool {
	if i == nil {
		return true
	}

	if reflect.TypeOf(i).Kind() == reflect.Ptr {
		return reflect.ValueOf(i).IsNil()
	}

	return false
}

// Read data from connection.
func (cm *ConnManager) Read(b []byte) (n int, err error) {
	cm.ConnectionMtx.RLock()
	defer cm.ConnectionMtx.RUnlock()
	return cm.Connection.Read(b)
}

// Write data to connection.
func (cm *ConnManager) Write(b []byte) (n int, err error) {
	cm.ConnectionMtx.RLock()
	defer cm.ConnectionMtx.RUnlock()
	return cm.Connection.Write(b)
}

// Close current connection.
func (cm *ConnManager) Close() error {
	cm.ConnectionMtx.RLock()
	defer cm.ConnectionMtx.RUnlock()
	return cm.Connection.Close()
}

// LocalAddr gets local address of connection.
func (cm *ConnManager) LocalAddr() net.Addr {
	cm.ConnectionMtx.RLock()
	defer cm.ConnectionMtx.RUnlock()
	return cm.Connection.LocalAddr()
}

// RemoteAddr gets remote address of connection.
func (cm *ConnManager) RemoteAddr() net.Addr {
	cm.ConnectionMtx.RLock()
	defer cm.ConnectionMtx.RUnlock()
	return cm.Connection.RemoteAddr()
}

// SetDeadline sets a deadline for current connection.
func (cm *ConnManager) SetDeadline(t time.Time) error {
	cm.ConnectionMtx.RLock()
	defer cm.ConnectionMtx.RUnlock()
	return cm.Connection.SetDeadline(t)
}

// SetReadDeadline sets read deadline for current connection.
func (cm *ConnManager) SetReadDeadline(t time.Time) error {
	cm.ConnectionMtx.RLock()
	defer cm.ConnectionMtx.RUnlock()
	return cm.Connection.SetReadDeadline(t)
}

// SetWriteDeadline sets write deadline for current connection.
func (cm *ConnManager) SetWriteDeadline(t time.Time) error {
	cm.ConnectionMtx.RLock()
	defer cm.ConnectionMtx.RUnlock()
	return cm.Connection.SetWriteDeadline(t)
}
