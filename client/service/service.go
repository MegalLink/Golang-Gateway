package service

import (
	"context"
	"encoding/json"
	"fmt"
	"megalink/gateway/client/channels"
	"megalink/gateway/client/connection"
	"megalink/gateway/client/types"
	"megalink/gateway/client/utils"
	"megalink/gateway/logger"
	"megalink/gateway/shared"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Service struct {
	Connection connection.IConnManager
	Logger     logger.IFastLogger
	Channel    *channels.ChannelStruct[*shared.Transaction]
}

func (sv *Service) TransactionService(c *gin.Context) {
	var requestBody types.ClientRequest
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		sv.Logger.Error("Error al decodificar el body", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Error al procesar la solicitud: " + err.Error(),
		})
		return
	}

	if err := sv.validateRequest(&requestBody); err != nil {
		sv.Logger.Error("Error de validación", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Error de validación: " + err.Error(),
		})
		return
	}

	res, err := sv.sendMessage(sv.getTransactionRequest(&requestBody))
	if err != nil {
		sv.Logger.Error("TransactionService", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error interno del servidor",
		})
		return
	}

	sv.Logger.Info("Transaction Service Response", res)

	c.JSON(http.StatusOK, res)
}

func (sv *Service) getTransactionRequest(requestBody *types.ClientRequest) *shared.Transaction {
	return &shared.Transaction{
		MTI: requestBody.TransactionType,
		F2:  requestBody.Card.Number,
		F3:  fmt.Sprintf("%s%s", requestBody.Card.ExpiryYear, requestBody.Card.ExpiryMonth),
		F4:  requestBody.Amount,
		F12: utils.GetTimeField(requestBody.Timezone),
		F13: utils.GetDateField(requestBody.Timezone),
	}
}

func (sv *Service) validateRequest(requestBody *types.ClientRequest) error {
	if requestBody.TransactionReference == "" {
		return fmt.Errorf("transaction_reference no proporcionado")
	}
	if requestBody.Amount == "" {
		return fmt.Errorf("amount no proporcionado")
	}
	if requestBody.Card.Number == "" {
		return fmt.Errorf("card no proporcionado")
	}

	return nil
}

func (sv *Service) sendMessage(req *shared.Transaction) (*shared.Transaction, error) {
	idChannel := fmt.Sprintf("%s%s", req.F12, req.F13)
	ch := sv.Channel.Init(idChannel)
	//TODO: change this time to env var
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
		return &shared.Transaction{F39: "TIMEOUT"}, nil
	}
}
