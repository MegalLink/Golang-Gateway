package types

type ClientRequest struct {
	TransactionReference string `json:"transaction_reference"`
	Card                 Card   `json:"card"`
	Amount               string `json:"amount"`
	TransactionType      string `json:"transaction_type"`
	Timezone             string `json:"timezone"`
}

type Card struct {
	Number      string `json:"number"`
	ExpiryYear  string `json:"expiry_year"`
	ExpiryMonth string `json:"expiry_month"`
}

type EnvVars struct {
	GinServerAdress              string
	FranchiseConnectionAdress    string
	ShowEcho                     bool
	ShowHeartBeat                bool
	HeartSendBeatIntervalSeconds int
	HeartBeatResponseWaitSeconds int
}
