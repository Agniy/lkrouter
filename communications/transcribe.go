package communications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"lkrouter/config"
	"lkrouter/pkg/awslogs"
	"net/http"
)

type TranscribeReq struct {
	Room          string `json:"room"`
	Lang          string `json:"lang"`
	TrackId       string `json:"trackID"`
	ParticipantId string `json:"participantId"`
	MsgId         string `json:"msgId"`
}

func (tr *TranscribeReq) RemoveMsgAction() (*TranscribeReq, error) {
	cfg := config.GetConfig()
	jsonData, err := json.Marshal(tr)
	if err != nil {
		awslogs.AddSLog(map[string]string{
			"func":    "RemoveMsgAction",
			"message": fmt.Sprintf("Error marshaling TranscribeReq: %v", err),
			"type":    awslogs.MsgTypeError,
			"room":    tr.Room,
		})
	}

	resp, err := http.Post(cfg.TranscribeAddr+"remove_msg", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		awslogs.AddSLog(map[string]string{
			"func":    "RemoveMsgAction",
			"message": fmt.Sprintf("Error in http.Post: %v, to url: %s", err, cfg.TranscribeAddr+"remove_msg"),
			"type":    awslogs.MsgTypeError,
			"room":    tr.Room})
	}

	defer resp.Body.Close()

	var transcriberResponse TranscribeReq
	err = json.NewDecoder(resp.Body).Decode(&transcriberResponse)
	if err != nil {

		awslogs.AddSLog(map[string]string{
			"func":    "TranscriberAction",
			"message": fmt.Sprintf("Error decoding response body: %v", err),
			"type":    awslogs.MsgTypeError,
			"room":    tr.Room,
		})

		return nil, err
	}

	return &transcriberResponse, nil
}

func (tr *TranscribeReq) TranscriberAction(action string) (*TranscribeReq, error) {
	cfg := config.GetConfig()
	jsonData, err := json.Marshal(tr)
	if err != nil {
		awslogs.AddSLog(map[string]string{
			"func":    "TranscriberAction",
			"message": fmt.Sprintf("Error marshaling TranscribeReq: %v", err),
			"type":    awslogs.MsgTypeError,
			"room":    tr.Room,
		})
	}

	resp, err := http.Post(cfg.TranscribeAddr+action, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		awslogs.AddSLog(map[string]string{
			"func":    "TranscriberAction",
			"message": fmt.Sprintf("Error in http.Post: %v, to url: %s", err, cfg.TranscribeAddr+action),
			"type":    awslogs.MsgTypeError,
			"room":    tr.Room,
		})
		return nil, err
	}
	defer resp.Body.Close()

	var transcriberResponse TranscribeReq
	err = json.NewDecoder(resp.Body).Decode(&transcriberResponse)
	if err != nil {

		awslogs.AddSLog(map[string]string{
			"func":    "TranscriberAction",
			"message": fmt.Sprintf("Error decoding response body: %v", err),
			"type":    awslogs.MsgTypeError,
			"room":    tr.Room,
		})

		return nil, err
	}

	return &transcriberResponse, nil
}
