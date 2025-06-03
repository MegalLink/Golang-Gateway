package sign

import (
	"encoding/json"
	"io"
	"megalink/gateway/client/types"
	"megalink/gateway/shared"
)

type (
	// ISignService is the interface of the Sign service.
	ISignService interface {
		SendSignOn(writer io.Writer) error
	}

	// SignService manage sending SignOn and SignOff messages to the franchise.
	SignService struct {
		EnvVars *types.EnvVars
	}
)

// NewSignService is the provider for new SignService.
func NewSignService(conf *types.EnvVars) ISignService {
	return &SignService{
		EnvVars: conf,
	}
}

// SendSignOn sends a SignOn request to the franchise.
func (sh *SignService) SendSignOn(writer io.Writer) error {
	signOnData := &shared.Transaction{
		MTI: "0200",
		F12: "",
		F13: "",
	}
	return sh.sendMessage(signOnData, writer)
}

func (sh *SignService) sendMessage(
	signData *shared.Transaction,
	writer io.Writer,
) error {
	// Encode heartbeat request to JSON
	requestBytes, err := json.Marshal(signData)
	if err != nil {
		return err
	}

	if _, err := writer.Write(requestBytes); err != nil {
		return err
	}

	return nil
}
