package handler

import (
	"context"
	"fmt"
	"io"
	"megalink/gateway/shared"
)

type (
	// ErrorHandler defines error handling methods.
	ErrorHandler interface {
		HandleMessageError(next MessageHandlerFunc) MessageHandlerFunc
		HandleError(ctx context.Context, err error)
	}
	// ListenerErrorHandler defines custom error handling logic to notify errors to Rollbar.
	ListenerErrorHandler struct {
	}
)

// NewErrorHandler provides an ErrorHandler.
func NewErrorHandler() ErrorHandler {
	return &ListenerErrorHandler{}
}

// HandleError notifies an error to rollbar.
// Add complex error handling logic as required.
func (eh *ListenerErrorHandler) HandleError(ctx context.Context, err error) {
	fmt.Printf("\nHandleError| %v", err)
}

// HandleMessageError notifies a message handler error to rollbar.
func (eh *ListenerErrorHandler) HandleMessageError(next MessageHandlerFunc) MessageHandlerFunc {
	return func(writer io.ReadWriter, data *shared.Transaction) error {
		err := next(writer, data)
		if err != nil {
			fmt.Printf("\nHandleMessageError| %v", err)
		}
		// return original err
		return err
	}
}
