package domain

type LKSttMsg struct {
	Type    string       `json:"type"`
	Payload SttActionMsg `json:"payload"`
}

type SttActionMsg struct {
	Enabled bool   `json:"enabled"`
	Lang    string `json:"lang"`
}
