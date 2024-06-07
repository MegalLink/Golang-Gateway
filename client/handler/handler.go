package handler

import (
	"io"
	"megalink/gateway/client/types"
	"net"
)

// MessageHandlerFunc it's the function which performs the actual processing of the actual data going through every
// handler.
type MessageHandlerFunc func(io.ReadWriter, *types.ServerResponse) error

// INetConn provides methods to deal with net.Conn.
type INetConn interface {
	net.Conn
}

// IAddress provides methods to deal with net.Addr.
type IAddress interface {
	net.Addr
}

// IOReaderWriter provides read/write methods to deal with io.ReadWriter.
type IOReaderWriter interface {
	io.ReadWriter
}

// ListenerHandlerFunc it's a closure which returns a MessageHandlerFunc.
type ListenerHandlerFunc func(next MessageHandlerFunc) MessageHandlerFunc
