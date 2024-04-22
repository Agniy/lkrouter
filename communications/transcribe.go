package communications

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"net/http"
)

type TranscribeReq struct {
	Id               string   `json:"id"`
	Uid              string   `json:"uid"`
	Room             string   `json:"room"`
	Lang             string   `json:"lang"`
	LangText         string   `json:"lang_text"`
	LangAlternatives []string `json:"langAlternatives"`
	Action           string   `json:"-"`
}

type BaseAddrResponse struct {
	Port int `json:"port"`
}

func GetTranscribePort(addrUri string, transcReq *TranscribeReq) (int, error) {

	logger := logrus.New()

	logger.Infof("GetTranscribePort start, by url %s", addrUri)
	data, err := json.Marshal(transcReq)
	if err != nil {
		return 0, err
	}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	resp, err := http.Post(addrUri, "application/json",
		bytes.NewBuffer(data))
	if err != nil {
		logger.Errorf("http.Post %v", err)
		return 0, err
	}

	var res BaseAddrResponse
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		logger.Errorf("BaseAddrResponse decode %v", err)
		return 0, err
	}
	logger.Infof("got transcribe port %d", res.Port)
	return res.Port, nil
}
