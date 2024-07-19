package service

import (
	"context"
	"encoding/json"
	"megalink/gateway/client/channels"
	"megalink/gateway/client/connection"
	"megalink/gateway/client/types"
	"megalink/gateway/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Service struct {
	Connection connection.IConnManager
	Logger     logger.IFastLogger
	Channel    *channels.ChannelStruct[*types.ServerResponse]
}

func (sv *Service) TransactionService(c *gin.Context) {
	request := types.ServerRequest{MessageType: "TRANSACTION", ServerResponse: "OK", RequestID: "TEST"}
	sv.Logger.Info("Transaction Service Init Request", request)
	res, err := sv.sendMessage(&request)
	if err != nil {
		sv.Logger.Error("TransactionService", err)
		c.JSON(http.StatusInternalServerError, err)
	}

	sv.Logger.Info("Transaction Service Response", res)
	c.JSON(http.StatusOK, res)
}

func (sv *Service) sendMessage(req *types.ServerRequest) (*types.ServerResponse, error) {
	idChannel := req.RequestID
	ch := sv.Channel.Init(idChannel)
	const timeOutChannel = 20 * time.Second
	ctxTimeOut, cancel := context.WithTimeout(context.Background(), timeOutChannel)
	defer func() {
		close(ch)
		sv.Channel.Delete(idChannel)
		cancel()
	}()

	requestBytes, _ := json.Marshal(req)
	_, err := sv.Connection.Write(requestBytes)
	if err != nil {
		sv.Logger.Error("Error sending transaction", err)
	}

	select {
	case response := <-ch:
		re := response.Resp
		sv.Logger.Info("Service response", re)
		return re, nil

	case <-ctxTimeOut.Done():
		return &types.ServerResponse{MessageType: "TRANSACTION", ServerResponse: "TIMEOUT"}, nil
	}
}
