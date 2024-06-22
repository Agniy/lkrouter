package domain

type LKSttMsg struct {
	Type    string       `json:"type"`
	Payload SttActionMsg `json:"payload"`
}

type SttActionMsg struct {
	Action string `json:"action"`
	Lang   string `json:"lang"`
}
