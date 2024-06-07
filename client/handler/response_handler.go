package handler

import (
	"context"
	"fmt"
	"io"
	"megalink/gateway/client/channels"
	"megalink/gateway/client/types"
)

type (
	// ResponseHandler defines handling methods.
	ResponseHandler interface {
		HandleMessageResponse(next MessageHandlerFunc) MessageHandlerFunc
	}
	// ListenerResponseHandler defines response handling logic.
	ListenerResponseHandler struct {
		// Ctx handles context.
		Ctx     context.Context
		Channel *channels.ChannelStruct[*types.ServerResponse]
	}
)

// NewResponseHandler provides an ResponseHandler.
func NewResponseHandler(ctx context.Context, channel *channels.ChannelStruct[*types.ServerResponse]) ResponseHandler {
	return &ListenerResponseHandler{
		Ctx:     ctx,
		Channel: channel,
	}
}

// HandleMessageResponse handles message response.
func (lrh *ListenerResponseHandler) HandleMessageResponse(_ MessageHandlerFunc) MessageHandlerFunc {
	return func(_ io.ReadWriter, response *types.ServerResponse) error {
		fmt.Println("HandleMessageResponse", response)
		// create a channel
		idCh := response.RequestID
		lrh.Channel.Set(channels.CHMessageFields[*types.ServerResponse]{
			Resp: response,
			ID:   idCh,
		})

		return nil
	}
}
