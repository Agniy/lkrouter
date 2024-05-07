package communications

import (
	"bytes"
	"encoding/json"
	"lkrouter/config"
	"log"
	"net/http"
)

type TranscribeReq struct {
	Room          string `json:"room"`
	Lang          string `json:"lang"`
	TrackId       string `json:"trackID"`
	ParticipantId string `json:"participantId"`
}

func (tr *TranscribeReq) TranscriberAction(action string) (TranscribeReq, error) {
	cfg := config.GetConfig()
	jsonData, err := json.Marshal(tr)
	if err != nil {
		log.Fatalf("Error occurred during marshaling. Err: %s", err.Error())
	}

	resp, err := http.Post(cfg.TranscribeAddr+action, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Error occurred during request. Err: %s", err.Error())
	}
	defer resp.Body.Close()

	var transcriberResponse TranscribeReq
	err = json.NewDecoder(resp.Body).Decode(&transcriberResponse)
	if err != nil {
		return TranscribeReq{}, err
	}

	return transcriberResponse, nil
}
