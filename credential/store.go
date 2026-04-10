package credential

import (
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	ServiceName = "ThreatIntelFeedWizard"
	AccountKey  = "api-key"
	AccountReg  = "region"
	AccountSync = "last-sync"
)

// Store abstracts secure credential persistence.
type Store interface {
	SaveAPIKey(key string) error
	LoadAPIKey() (string, error)
	SaveRegion(region string) error
	LoadRegion() (string, error)
	SaveLastSync(ts string) error
	LoadLastSync() (string, error)
}

// KeyringStore implements Store using the OS-native keyring (macOS Keychain / Windows Credential Manager).
type KeyringStore struct{}

func (k *KeyringStore) SaveAPIKey(key string) error {
	if err := keyring.Set(ServiceName, AccountKey, key); err != nil {
		return fmt.Errorf("save api key: %w", err)
	}
	return nil
}

func (k *KeyringStore) LoadAPIKey() (string, error) {
	val, err := keyring.Get(ServiceName, AccountKey)
	if err != nil {
		return "", fmt.Errorf("load api key: %w", err)
	}
	return val, nil
}

func (k *KeyringStore) SaveRegion(region string) error {
	if err := keyring.Set(ServiceName, AccountReg, region); err != nil {
		return fmt.Errorf("save region: %w", err)
	}
	return nil
}

func (k *KeyringStore) LoadRegion() (string, error) {
	val, err := keyring.Get(ServiceName, AccountReg)
	if err != nil {
		return "", fmt.Errorf("load region: %w", err)
	}
	return val, nil
}

func (k *KeyringStore) SaveLastSync(ts string) error {
	if err := keyring.Set(ServiceName, AccountSync, ts); err != nil {
		return fmt.Errorf("save last sync: %w", err)
	}
	return nil
}

func (k *KeyringStore) LoadLastSync() (string, error) {
	val, err := keyring.Get(ServiceName, AccountSync)
	if err != nil {
		return "", fmt.Errorf("load last sync: %w", err)
	}
	return val, nil
}
