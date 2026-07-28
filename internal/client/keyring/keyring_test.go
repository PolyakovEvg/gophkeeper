package keyring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyring_SetAndGet(t *testing.T) {
	k := New()
	err := k.SetMasterPassword("testuser", "secret123")
	if err != nil {
		t.Skipf("keyring not available: %v", err)
	}

	password, err := k.GetMasterPassword("testuser")
	assert.NoError(t, err)
	assert.Equal(t, "secret123", password)
}

func TestKeyring_GetNotFound(t *testing.T) {
	k := New()
	_, err := k.GetMasterPassword("nonexistent")
	assert.Error(t, err)
}

func TestKeyring_Delete(t *testing.T) {
	k := New()
	err := k.SetMasterPassword("deleteuser", "secret")
	if err != nil {
		t.Skipf("keyring not available: %v", err)
	}

	err = k.DeleteMasterPassword("deleteuser")
	assert.NoError(t, err)

	_, err = k.GetMasterPassword("deleteuser")
	assert.Error(t, err)
}
