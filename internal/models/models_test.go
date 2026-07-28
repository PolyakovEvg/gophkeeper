package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestItemType_Valid(t *testing.T) {
	tests := []struct {
		name  string
		item  ItemType
		valid bool
	}{
		{"credentials", ItemCredentials, true},
		{"text", ItemText, true},
		{"binary", ItemBinary, true},
		{"bank_card", ItemBankCard, true},
		{"otp", ItemOTP, true},
		{"invalid", ItemType("invalid"), false},
		{"empty", ItemType(""), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.item.Valid())
		})
	}
}

func TestUser(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	user := User{
		ID:           id,
		Login:        "testuser",
		PasswordHash: "hash",
		CreatedAt:    now,
	}
	assert.Equal(t, id, user.ID)
	assert.Equal(t, "testuser", user.Login)
	assert.Equal(t, "hash", user.PasswordHash)
	assert.Equal(t, now, user.CreatedAt)
}

func TestVaultItem(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	now := time.Now()
	item := VaultItem{
		ID:        id,
		UserID:    userID,
		Version:   1,
		UpdatedAt: now,
		Deleted:   false,
		Payload:   []byte("test"),
	}
	assert.Equal(t, id, item.ID)
	assert.Equal(t, userID, item.UserID)
	assert.Equal(t, int64(1), item.Version)
	assert.Equal(t, now, item.UpdatedAt)
	assert.False(t, item.Deleted)
	assert.Equal(t, []byte("test"), item.Payload)
}

func TestCredentialsData(t *testing.T) {
	data := CredentialsData{
		Login:    "user",
		Password: "pass",
	}
	assert.Equal(t, "user", data.Login)
	assert.Equal(t, "pass", data.Password)
}

func TestTextData(t *testing.T) {
	data := TextData{Content: "some text"}
	assert.Equal(t, "some text", data.Content)
}

func TestBinaryData(t *testing.T) {
	data := BinaryData{
		Content: []byte{1, 2, 3},
		Mime:    "application/pdf",
	}
	assert.Equal(t, []byte{1, 2, 3}, data.Content)
	assert.Equal(t, "application/pdf", data.Mime)
}

func TestBankCardData(t *testing.T) {
	data := BankCardData{
		Number: "4111111111111111",
		Holder: "Test User",
		Expiry: "12/25",
		CVV:    "123",
		Bank:   "Test Bank",
	}
	assert.Equal(t, "4111111111111111", data.Number)
	assert.Equal(t, "Test User", data.Holder)
	assert.Equal(t, "12/25", data.Expiry)
	assert.Equal(t, "123", data.CVV)
	assert.Equal(t, "Test Bank", data.Bank)
}

func TestOTPData(t *testing.T) {
	data := OTPData{
		Secret:    "JBSWY3DPEHPK3PXP",
		Issuer:    "Test App",
		Account:   "user@example.com",
		Algorithm: "SHA1",
		Digits:    6,
		Period:    30,
		Type:      "totp",
	}
	assert.Equal(t, "JBSWY3DPEHPK3PXP", data.Secret)
	assert.Equal(t, "Test App", data.Issuer)
	assert.Equal(t, "user@example.com", data.Account)
	assert.Equal(t, "SHA1", data.Algorithm)
	assert.Equal(t, 6, data.Digits)
	assert.Equal(t, 30, data.Period)
	assert.Equal(t, "totp", data.Type)
}

func TestItemPayload(t *testing.T) {
	creds := &CredentialsData{Login: "user", Password: "pass"}
	payload := ItemPayload{
		Type:        ItemCredentials,
		Title:       "My Login",
		Metadata:    map[string]string{"url": "https://example.com"},
		Credentials: creds,
	}
	assert.Equal(t, ItemCredentials, payload.Type)
	assert.Equal(t, "My Login", payload.Title)
	assert.Equal(t, "https://example.com", payload.Metadata["url"])
	assert.Equal(t, creds, payload.Credentials)
}

func TestLocalItem(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	payload := ItemPayload{
		Type:  ItemText,
		Title: "Note",
	}
	item := LocalItem{
		ID:        id,
		Version:   1,
		UpdatedAt: now,
		Deleted:   false,
		Dirty:     true,
		Payload:   payload,
	}
	assert.Equal(t, id, item.ID)
	assert.Equal(t, int64(1), item.Version)
	assert.Equal(t, now, item.UpdatedAt)
	assert.False(t, item.Deleted)
	assert.True(t, item.Dirty)
	assert.Equal(t, payload, item.Payload)
}
