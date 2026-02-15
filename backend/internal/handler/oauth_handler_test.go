package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const testBotToken = "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"

func computeValidHash(data map[string]interface{}, botToken string) string {
	var parts []string
	for k, v := range data {
		if k == "hash" {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%s", k, formatTelegramValue(v)))
	}
	sort.Strings(parts)
	dataCheckString := strings.Join(parts, "\n")

	secretKey := sha256.Sum256([]byte(botToken))
	mac := hmac.New(sha256.New, secretKey[:])
	mac.Write([]byte(dataCheckString))
	return hex.EncodeToString(mac.Sum(nil))
}

func buildTelegramData(botToken string) map[string]interface{} {
	data := map[string]interface{}{
		"id":         float64(123456789),
		"first_name": "John",
		"last_name":  "Doe",
		"username":   "johndoe",
		"photo_url":  "https://t.me/i/userpic/320/johndoe.jpg",
		"auth_date":  float64(time.Now().Unix()),
	}
	data["hash"] = computeValidHash(data, botToken)
	return data
}

func TestVerifyTelegramAuth_ValidData(t *testing.T) {
	h := &OAuthHandler{telegramBotToken: testBotToken}

	data := buildTelegramData(testBotToken)

	assert.True(t, h.verifyTelegramAuth(data))
}

func TestVerifyTelegramAuth_InvalidHash(t *testing.T) {
	h := &OAuthHandler{telegramBotToken: testBotToken}

	data := buildTelegramData(testBotToken)
	data["hash"] = "0000000000000000000000000000000000000000000000000000000000000000"

	assert.False(t, h.verifyTelegramAuth(data))
}

func TestVerifyTelegramAuth_MissingHash(t *testing.T) {
	h := &OAuthHandler{telegramBotToken: testBotToken}

	data := buildTelegramData(testBotToken)
	delete(data, "hash")

	assert.False(t, h.verifyTelegramAuth(data))
}

func TestVerifyTelegramAuth_ExpiredAuthDate(t *testing.T) {
	h := &OAuthHandler{telegramBotToken: testBotToken}

	data := map[string]interface{}{
		"id":         float64(123456789),
		"first_name": "John",
		"username":   "johndoe",
		"auth_date":  float64(time.Now().Add(-25 * time.Hour).Unix()),
	}
	data["hash"] = computeValidHash(data, testBotToken)

	assert.False(t, h.verifyTelegramAuth(data))
}

func TestVerifyTelegramAuth_EmptyBotToken(t *testing.T) {
	h := &OAuthHandler{telegramBotToken: ""}

	data := buildTelegramData("")

	assert.False(t, h.verifyTelegramAuth(data))
}

func TestVerifyTelegramAuth_TamperedField(t *testing.T) {
	h := &OAuthHandler{telegramBotToken: testBotToken}

	data := buildTelegramData(testBotToken)
	data["first_name"] = "Hacker"

	assert.False(t, h.verifyTelegramAuth(data))
}

func TestVerifyTelegramAuth_MissingAuthDate(t *testing.T) {
	h := &OAuthHandler{telegramBotToken: testBotToken}

	data := map[string]interface{}{
		"id":         float64(123456789),
		"first_name": "John",
		"username":   "johndoe",
	}
	data["hash"] = computeValidHash(data, testBotToken)

	assert.False(t, h.verifyTelegramAuth(data))
}

func TestVerifyTelegramAuth_FutureAuthDate(t *testing.T) {
	h := &OAuthHandler{telegramBotToken: testBotToken}

	data := map[string]interface{}{
		"id":         float64(123456789),
		"first_name": "John",
		"username":   "johndoe",
		"auth_date":  float64(time.Now().Add(1 * time.Hour).Unix()),
	}
	data["hash"] = computeValidHash(data, testBotToken)

	assert.False(t, h.verifyTelegramAuth(data))
}

func TestVerifyTelegramAuth_AuthDateNotNumber(t *testing.T) {
	h := &OAuthHandler{telegramBotToken: testBotToken}

	data := map[string]interface{}{
		"id":         float64(123456789),
		"first_name": "John",
		"username":   "johndoe",
		"auth_date":  "not-a-number",
	}
	data["hash"] = computeValidHash(data, testBotToken)

	assert.False(t, h.verifyTelegramAuth(data))
}
