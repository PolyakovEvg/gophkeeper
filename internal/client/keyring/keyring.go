// Package keyring provides secure storage for secrets using OS keychain.
package keyring

import (
	"errors"

	"github.com/zalando/go-keyring"
)

const (
	serviceName       = "gophkeeper"
	tokenKey          = "token"
	masterPasswordKey = "master_password"
)

// Keyring wraps OS keychain operations.
type Keyring struct{}

// New creates a new Keyring.
func New() *Keyring {
	return &Keyring{}
}

// SetToken stores the auth token in the keychain.
func (k *Keyring) SetToken(login, token string) error {
	return keyring.Set(serviceName, tokenKey+"_"+login, token)
}

// GetToken retrieves the auth token from the keychain.
func (k *Keyring) GetToken(login string) (string, error) {
	return keyring.Get(serviceName, tokenKey+"_"+login)
}

// DeleteToken removes the auth token from the keychain.
func (k *Keyring) DeleteToken(login string) error {
	return keyring.Delete(serviceName, tokenKey+"_"+login)
}

// SetMasterPassword stores the master password in the keychain.
func (k *Keyring) SetMasterPassword(login, password string) error {
	return keyring.Set(serviceName, masterPasswordKey+"_"+login, password)
}

// GetMasterPassword retrieves the master password from the keychain.
func (k *Keyring) GetMasterPassword(login string) (string, error) {
	return keyring.Get(serviceName, masterPasswordKey+"_"+login)
}

// DeleteMasterPassword removes the master password from the keychain.
func (k *Keyring) DeleteMasterPassword(login string) error {
	return keyring.Delete(serviceName, masterPasswordKey+"_"+login)
}

// ClearSession removes all session data from the keychain.
func (k *Keyring) ClearSession(login string) error {
	var errs []error
	if err := k.DeleteToken(login); err != nil && !errors.Is(err, keyring.ErrNotFound) {
		errs = append(errs, err)
	}
	if err := k.DeleteMasterPassword(login); err != nil && !errors.Is(err, keyring.ErrNotFound) {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}
