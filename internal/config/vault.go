package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

const (
	vaultMagic   = "ONE9S_VAULT_V1"
	vaultSaltLen = 16
	vaultIVLen   = 12
	vaultKeyLen  = 32
	pbkdf2Iter   = 100000
)

// IsVaultFile checks if data starts with the vault magic header.
func IsVaultFile(data []byte) bool {
	return len(data) >= len(vaultMagic) && string(data[:len(vaultMagic)]) == vaultMagic
}

// Encrypt encrypts plaintext config data with a password.
func Encrypt(plaintext []byte, password string) ([]byte, error) {
	salt := make([]byte, vaultSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("generate salt: %w", err)
	}

	iv := make([]byte, vaultIVLen)
	if _, err := rand.Read(iv); err != nil {
		return nil, fmt.Errorf("generate IV: %w", err)
	}

	key := deriveKey(password, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	ciphertext := gcm.Seal(nil, iv, plaintext, nil)

	// Build vault: magic + salt + IV + ciphertext
	result := make([]byte, 0, len(vaultMagic)+vaultSaltLen+vaultIVLen+len(ciphertext))
	result = append(result, []byte(vaultMagic)...)
	result = append(result, salt...)
	result = append(result, iv...)
	result = append(result, ciphertext...)

	return result, nil
}

// Decrypt decrypts vault data with a password.
func Decrypt(data []byte, password string) ([]byte, error) {
	if !IsVaultFile(data) {
		return nil, fmt.Errorf("not a valid vault file")
	}

	offset := len(vaultMagic)
	salt := data[offset : offset+vaultSaltLen]
	offset += vaultSaltLen
	iv := data[offset : offset+vaultIVLen]
	offset += vaultIVLen
	ciphertext := data[offset:]

	key := deriveKey(password, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	plaintext, err := gcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed (wrong password?)")
	}

	return plaintext, nil
}

func deriveKey(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, pbkdf2Iter, vaultKeyLen, sha256.New)
}

// VaultPath returns the default vault file path.
func VaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".one9s", "config.vault")
}

// InitVault creates a new vault file with credentials.
func InitVault(path, password, user, pass, endpoint string) error {
	content := fmt.Sprintf("ONE_XMLRPC=%s\nONE_AUTH=%s:%s\n", endpoint, user, pass)
	encrypted, err := Encrypt([]byte(content), password)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	dir := strings.TrimSuffix(path, string(os.PathSeparator)+"config.vault")
	os.MkdirAll(dir, 0700)

	if err := os.WriteFile(path, encrypted, 0600); err != nil {
		return fmt.Errorf("write vault: %w", err)
	}
	return nil
}

// EncryptConfig encrypts an existing plain config file.
func EncryptConfig(vaultPath, configPath, password string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	encrypted, err := Encrypt(data, password)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	dir := strings.TrimSuffix(vaultPath, string(os.PathSeparator)+"config.vault")
	os.MkdirAll(dir, 0700)

	if err := os.WriteFile(vaultPath, encrypted, 0600); err != nil {
		return fmt.Errorf("write vault: %w", err)
	}
	return nil
}

// DecryptVault reads and decrypts a vault file.
func DecryptVault(vaultPath, password string) (string, error) {
	data, err := os.ReadFile(vaultPath)
	if err != nil {
		return "", fmt.Errorf("read vault: %w", err)
	}

	plaintext, err := Decrypt(data, password)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// VaultPassword gets the vault password from env var or returns empty.
func VaultPassword() string {
	return os.Getenv("ONE_VAULT_PASS")
}
