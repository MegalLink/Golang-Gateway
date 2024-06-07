package listener

import (
	"io"
	"megalink/gateway/client/handler"
	"megalink/gateway/client/types"
)

// ListenerChain is a set of handler functions which will be applied to the messages flowing through that connection.
type ListenerChain struct {
	handlers []handler.ListenerHandlerFunc
}

// BuildChain gets the first handler function which will be used by a listener.
func (ch *ListenerChain) BuildChain() handler.MessageHandlerFunc {
	var head handler.MessageHandlerFunc

	for i := len(ch.handlers) - 1; i >= 0; i-- {
		if head == nil {
			head = ch.handlers[i](doNothing)
			continue
		}
		head = ch.handlers[i](head)
	}

	// deletes current handlers
	ch.handlers = nil
	return head
}

func doNothing(io.ReadWriter, *types.ServerResponse) error { return nil }

// AddHandler appends a new handler to the handlers set.
func (ch *ListenerChain) AddHandler(h handler.ListenerHandlerFunc) *ListenerChain {
	ch.handlers = append(ch.handlers, h)
	return ch
}
