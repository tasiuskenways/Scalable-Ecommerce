package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// GenerateRSAKeyPair generates a new RSA key pair
func GenerateRSAKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}
	return privateKey, &privateKey.PublicKey, nil
}

// SavePEMKey saves a key to a PEM file
func SavePEMKey(filename string, key *rsa.PrivateKey) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(key)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	if err := pem.Encode(file, privateKeyBlock); err != nil {
		return err
	}

	return nil
}

// SavePublicPEMKey saves a public key to a PEM file
func SavePublicPEMKey(filename string, pubkey *rsa.PublicKey) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(pubkey)
	if err != nil {
		return err
	}

	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	if err := pem.Encode(file, publicKeyBlock); err != nil {
		return err
	}

	return nil
}

// LoadPrivateKey loads a private key from a PEM file
func LoadPrivateKey(filename string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// LoadPublicKey loads a public key from a PEM file
func LoadPublicKey(filename string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("failed to decode PEM block containing public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		return nil, errors.New("key type is not RSA")
	}
}

// EncryptWithPublicKey encrypts data with RSA public key
func EncryptWithPublicKey(publicKey *rsa.PublicKey, data []byte) ([]byte, error) {
	hash := sha256.New()
	ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, publicKey, data, nil)
	if err != nil {
		return nil, err
	}
	return ciphertext, nil
}

// DecryptWithPrivateKey decrypts data with RSA private key
func DecryptWithPrivateKey(privateKey *rsa.PrivateKey, ciphertext []byte) ([]byte, error) {
	hash := sha256.New()
	plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, privateKey, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

// GenerateAESKey generates a new AES-256 key
func GenerateAESKey() ([]byte, error) {
	key := make([]byte, 32) // 32 bytes for AES-256
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}

// EncryptAES encrypts data using AES-GCM
func EncryptAES(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)
	return append(nonce, ciphertext...), nil
}

// GenerateAndSaveRSAKeyPair generates a new RSA key pair and saves them as PEM files in the specified directory
// Returns the paths to the generated private and public key files
func GenerateAndSaveRSAKeyPair(dirPath string) (privateKeyPath, publicKeyPath string, err error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(dirPath, 0700); err != nil {
		return "", "", fmt.Errorf("failed to create directory: %v", err)
	}

	// Generate RSA key pair
	privateKey, publicKey, err := GenerateRSAKeyPair()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate RSA key pair: %v", err)
	}

	// Define file paths
	privateKeyPath = filepath.Join(dirPath, "private_key.pem")
	publicKeyPath = filepath.Join(dirPath, "public_key.pem")

	// Save private key
	if err := SavePEMKey(privateKeyPath, privateKey); err != nil {
		return "", "", fmt.Errorf("failed to save private key: %v", err)
	}

	// Save public key
	if err := SavePublicPEMKey(publicKeyPath, publicKey); err != nil {
		// Clean up private key file if public key save fails
		_ = os.Remove(privateKeyPath)
		return "", "", fmt.Errorf("failed to save public key: %v", err)
	}

	return privateKeyPath, publicKeyPath, nil
}

// DecryptAES decrypts data using AES-GCM
func DecryptAES(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesgcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
