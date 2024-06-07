package connection

import (
	"fmt"
	"megalink/gateway/client/types"
	"net"
)

var (
	// NetDialerFn dialer fn to provide an insecure net.Conn.
	NetDialerFn DialerFn = net.Dial
)

type (
	// DialerFn function signature to connect with server over an insecure TCP connection.
	DialerFn func(network string, address string) (net.Conn, error)

	// IConnFactory is a net.Conn provider.
	IConnFactory interface {
		GetConnection() (net.Conn, error)
	}

	// ConnFactory deals with connection details to provide net.Conn per environment.
	ConnFactory struct {
		Cfg *types.EnvVars
	}
)

// NewConnFactory initializes a new IConnFactory.
func NewConnFactory(
	envCfg *types.EnvVars,
) IConnFactory {

	return &ConnFactory{
		Cfg: envCfg,
	}
}

// GetConnection creates a new net.Conn.
func (cf *ConnFactory) GetConnection() (net.Conn, error) {
	println("\nGet connection")

	net, err := cf.providePlainConnection()
	if err != nil {
		fmt.Printf("\n GetConnection |Connection failed %v", err)
	}

	return net, err
}

func (cf *ConnFactory) providePlainConnection() (net.Conn, error) {
	return cf.provideInsecureConnection()
}

// ProvideConnection provides a simple TCP connection.
func (cf *ConnFactory) provideInsecureConnection() (net.Conn, error) {
	address := cf.Cfg.FranchiseConnectionAdress

	fmt.Printf("\nTrying to establish an insecure connection with %s \n", address)

	return NetDialerFn("tcp", address)
}
