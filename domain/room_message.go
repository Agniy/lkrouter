package domain

type RoomActionMessage struct {
	Action string `json:"action"`
}

type RoomHttpNotification struct {
	MsgCode  string `json:"msgCode"`
	Type     string `json:"type"`
	Head     string `json:"head"`
	Msg      string `json:"msg"`
	Infinite bool   `json:"infinite"`
}
