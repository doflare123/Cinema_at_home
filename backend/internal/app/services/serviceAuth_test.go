package services

import (
	"cinema/internal/models/dto"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestVerifyTelegramAuthPayloadValid(t *testing.T) {
	botToken := "bot-token"
	req := dto.TelegramAuthRequest{
		TelegramID:       12345,
		TelegramUsername: "tester",
		DisplayName:      "Tester",
		AvatarURL:        "https://example.com/a.png",
		AuthDate:         time.Now().Unix(),
	}
	req.Hash = buildTelegramHash(req, botToken)

	if err := verifyTelegramAuthPayload(req, botToken); err != nil {
		t.Fatalf("expected payload to be valid, got %v", err)
	}
}

func TestVerifyTelegramAuthPayloadInvalidHash(t *testing.T) {
	botToken := "bot-token"
	req := dto.TelegramAuthRequest{
		TelegramID:       12345,
		TelegramUsername: "tester",
		DisplayName:      "Tester",
		AuthDate:         time.Now().Unix(),
		Hash:             "deadbeef",
	}

	if err := verifyTelegramAuthPayload(req, botToken); err == nil {
		t.Fatal("expected invalid hash error")
	}
}

func buildTelegramHash(req dto.TelegramAuthRequest, botToken string) string {
	dataMap := map[string]string{
		"auth_date":         fmt.Sprintf("%d", req.AuthDate),
		"display_name":      req.DisplayName,
		"telegram_id":       fmt.Sprintf("%d", req.TelegramID),
		"telegram_username": req.TelegramUsername,
		"avatar_url":        req.AvatarURL,
	}
	keys := make([]string, 0, len(dataMap))
	for k, v := range dataMap {
		if v == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	lines := make([]string, 0, len(keys))
	for _, k := range keys {
		lines = append(lines, k+"="+dataMap[k])
	}
	dataCheckString := strings.Join(lines, "\n")

	secret := sha256.Sum256([]byte(botToken))
	mac := hmac.New(sha256.New, secret[:])
	_, _ = mac.Write([]byte(dataCheckString))
	return hex.EncodeToString(mac.Sum(nil))
}
