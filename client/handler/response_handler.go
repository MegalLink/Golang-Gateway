package handler

import (
	"context"
	"fmt"
	"io"
	"megalink/gateway/client/channels"
	"megalink/gateway/shared"
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
		Channel *channels.ChannelStruct[*shared.Transaction]
	}
)

// NewResponseHandler provides an ResponseHandler.
func NewResponseHandler(ctx context.Context, channel *channels.ChannelStruct[*shared.Transaction]) ResponseHandler {
	return &ListenerResponseHandler{
		Ctx:     ctx,
		Channel: channel,
	}
}

// HandleMessageResponse handles message response.
func (lrh *ListenerResponseHandler) HandleMessageResponse(_ MessageHandlerFunc) MessageHandlerFunc {
	return func(_ io.ReadWriter, response *shared.Transaction) error {
		idCh := fmt.Sprintf("%s%s", response.F12, response.F13)
		lrh.Channel.Set(channels.CHMessageFields[*shared.Transaction]{
			Resp: response,
			ID:   idCh,
		})

		return nil
	}
}
