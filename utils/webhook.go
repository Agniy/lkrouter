package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
)

func SendWebhookData(data []byte, url, username, password string) error {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Handle the response here...

	return nil
}

func EncryptAuthData(username, password, timestamp string) string {
	authStr := fmt.Sprintf("%s:%s:%s", username, password, timestamp)
	hash := sha256.Sum256([]byte(authStr))
	return hex.EncodeToString(hash[:])
}
