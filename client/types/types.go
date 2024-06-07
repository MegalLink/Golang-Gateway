package types

type ServerRequest struct {
	RequestID      string `json:"requestID"`
	MessageType    string `json:"messageType"`
	ServerResponse string `json:"serverResponse"`
}

type ServerResponse struct {
	RequestID      string `json:"requestID"`
	MessageType    string `json:"messageType"`
	ServerResponse string `json:"serverResponse"`
}

type EnvVars struct {
	GinServerAdress              string
	FranchiseConnectionAdress    string
	ShowEcho                     bool
	HeartSendBeatIntervalSeconds int
	HeartBeatResponseWaitSeconds int
}
