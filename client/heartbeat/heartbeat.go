package heartbeat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"megalink/gateway/client/handler"
	"megalink/gateway/client/types"
	"megalink/gateway/client/utils"
	"megalink/gateway/logger"
	"megalink/gateway/shared"
	"sync/atomic"
	"time"
)

const (
	// Maximum heartbeat retries before performing a sign on.
	maxHeartbeatRetries = 3
	// DE39 echo test successfully response code.
	deEchoSuccessfully = "00"
	messageTypeEcho    = "ECHO"
)

var (
	// ErrHeartbeat error triggered if heartbeat is unable to receive responses.
	ErrHeartbeat = errors.New("heartbeat error")
)

// IHeartbeatService heartbeat service definition to handle echo test messages.
type IHeartbeatService interface {
	SendEchoTest(writer io.ReadWriter)
	HandleHeartBeatResponse(next handler.MessageHandlerFunc) handler.MessageHandlerFunc
	GetError() <-chan error
}

// HeartBeatService deals with heartbeat details of Datafast Connection.
type HeartBeatService struct {
	EchoTestResponse chan *shared.Transaction
	LastAlert        time.Time
	EchoRetries      uint64
	echoError        chan error
	EnvVars          *types.EnvVars
	WaitResponseTime time.Duration
	Logger           logger.IFastLogger
}

// NewHeartBeatService provides a new HeartBeatService with default config.
func NewHeartBeatService(envVars *types.EnvVars, logger logger.IFastLogger) IHeartbeatService {
	return &HeartBeatService{
		EchoTestResponse: make(chan *shared.Transaction),
		EchoRetries:      0,
		echoError:        make(chan error, 1),
		EnvVars:          envVars,
		WaitResponseTime: time.Duration(envVars.HeartBeatResponseWaitSeconds) * time.Second,
		Logger:           logger,
	}
}

// SendEchoTest send echo test messages through writer connection.
func (hb *HeartBeatService) SendEchoTest(writer io.ReadWriter) {
	if hb.EnvVars.ShowEcho {
		fmt.Printf("\nSendEchoTest ======= echo retries %d", atomic.LoadUint64(&hb.EchoRetries))
	}

	request := &shared.Transaction{
		MTI: messageTypeEcho,
		F12: utils.GetTimeField("UTC"),
		F13: utils.GetTimeField("UTC"),
	}
	// Encode heartbeat request to JSON
	requestBytes, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("\nSendEchoTest | Marshall err %v", err)
	}

	if hb.EnvVars.ShowEcho {
		fmt.Printf("\nSendEchoTest ======= sending %s", string(requestBytes))
	}

	if _, err = writer.Write(requestBytes); err != nil && hb.EnvVars.ShowEcho {
		fmt.Printf("\nSendEchoTest | Write err %v ", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), hb.WaitResponseTime)

	defer cancel()

	hb.checkEchoTestResponse(ctx)
}

func (hb *HeartBeatService) sendHeartBeatAlert(isContextDone bool) {
	hb.Logger.Warning("sendHeartBeatAlert", "isContextDone "+fmt.Sprint(isContextDone))

	atomic.AddUint64(&hb.EchoRetries, 1)
	if atomic.LoadUint64(&hb.EchoRetries) == maxHeartbeatRetries {
		atomic.SwapUint64(&hb.EchoRetries, 0)
		hb.Logger.Error("sendHeartBeatAlert", "Echo test failed 3 times "+fmt.Sprint(isContextDone))
		hb.sendHeartbeatError(ErrHeartbeat)
	}
}

func (hb *HeartBeatService) sendHeartbeatError(err error) {
	select {
	case hb.echoError <- err:
	default:
	}
}

func (hb *HeartBeatService) checkEchoTestResponse(ctx context.Context) {
	for {
		select {
		//This case is triggered if the context's timeout has expired for waiting response
		case <-ctx.Done():
			log.Println("checkEchoTestResponse context done")
			hb.sendHeartBeatAlert(true)
			return
		case res := <-hb.EchoTestResponse:
			hb.checkEchoResponse(res)
			return
		}
	}
}

func (hb *HeartBeatService) checkEchoResponse(res *shared.Transaction) {
	if hb.EnvVars.ShowEcho {
		fmt.Printf("\ncheckEchoResponse | response %v ", res)
	}

	if res.F39 != deEchoSuccessfully {
		hb.sendHeartBeatAlert(false)
		return
	}
	atomic.SwapUint64(&hb.EchoRetries, 0)
}

// HandleHeartBeatResponse handles echo test response from Datafast.
func (hb *HeartBeatService) HandleHeartBeatResponse(next handler.MessageHandlerFunc) handler.MessageHandlerFunc {
	return func(conn io.ReadWriter, response *shared.Transaction) error {
		// if is not type echo send to next handler
		if response.MTI != messageTypeEcho {
			return next(conn, response)
		}

		if hb.EnvVars.ShowEcho {
			fmt.Printf("\nHandleHeartBeatResponse | response %v ", response)
		}

		select {
		case hb.EchoTestResponse <- response:
		default:
		}

		return nil
	}
}

// GetError gets error notification channel of heartbeat.
func (hb *HeartBeatService) GetError() <-chan error {
	return hb.echoError
}
