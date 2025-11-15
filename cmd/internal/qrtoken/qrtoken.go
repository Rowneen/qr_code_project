package qrtoken

import (
	"encoding/json"
	"qr_code/internal/cipher"
	"time"
)

type QrToken struct {
	ID          int64  `json:"id"`
	Name        string `json:"nameLesson"`
	Date        string `json:"date"`
	Type        string `json:"type"`
	TeacherName string `json:"teacherName"`
	Created     int64  `json:"created"`
}

var qrTokenSecretKey = []byte("A3k9XpL2qR8mZbN7vGyJ4tHwE1cF6dS5")

func Generate(id int64, nameLesson, date, typeLes, teacherName string) (string, error) {
	token := map[string]interface{}{
		"id":          id,
		"nameLesson":  nameLesson,
		"date":        date,
		"type":        typeLes,
		"teacherName": teacherName,
		"created":     time.Now().Unix(),
	}

	return cipher.EncryptAES(token, qrTokenSecretKey)
}

func Parse(encrypted string) (*QrToken, error) {
	data, err := cipher.DecryptAES(encrypted, qrTokenSecretKey)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var token QrToken
	if err := json.Unmarshal(jsonData, &token); err != nil {
		return nil, err
	}
	// todo time check

	return &token, nil
}
