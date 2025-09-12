package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/tasiuskenways/scalable-ecommerce/crypto-service/internal/utils/crypto"
)

func main() {
	// Define command-line flags
	action := flag.String("action", "encrypt", "Action to perform: 'encrypt' or 'decrypt'")
	input := flag.String("input", "", "Input JSON string to encrypt/decrypt (e.g., '{\"key\":\"value\"}')")
	privateKeyPath := flag.String("private", "../keys/private.pem", "Path to private key")
	publicKeyPath := flag.String("public", "../keys/public.pem", "Path to public key")
	flag.Parse()

	switch *action {
	case "encrypt":
		// Parse the input as JSON
		var data interface{}
		if err := json.Unmarshal([]byte(*input), &data); err != nil {
			log.Fatalf("Failed to parse input as JSON: %v", err)
		}
		encryptMessage(*publicKeyPath, data)
	case "decrypt":
		result, err := decryptMessage(*privateKeyPath, *input)
		if err != nil {
			log.Fatalf("Decryption failed: %v", err)
		}

		// Pretty print the decrypted JSON
		jsonOutput, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			log.Fatalf("Failed to format decrypted data: %v", err)
		}
		fmt.Println("Decrypted data:")
		fmt.Println(string(jsonOutput))
	default:
		log.Fatalf("Invalid action: %s. Use 'encrypt' or 'decrypt'", *action)
	}
}

func encryptMessage(publicKeyPath string, data interface{}) {
	// Load public key
	publicKey, err := crypto.LoadPublicKey(publicKeyPath)
	if err != nil {
		log.Fatalf("Failed to load public key: %v", err)
	}

	// Convert data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Failed to marshal data to JSON: %v", err)
	}

	// Generate a new AES key
	aesKey, err := crypto.GenerateAESKey()
	if err != nil {
		log.Fatalf("Failed to generate AES key: %v", err)
	}

	// Encrypt the JSON data with AES
	ciphertext, err := crypto.EncryptAES(aesKey, jsonData)
	if err != nil {
		log.Fatalf("Failed to encrypt data: %v", err)
	}

	// Encrypt the AES key with RSA
	encryptedKey, err := crypto.EncryptWithPublicKey(publicKey, aesKey)
	if err != nil {
		log.Fatalf("Failed to encrypt AES key: %v", err)
	}

	// Output the encrypted key and data in a format that can be easily used for decryption
	encryptedKeyHex := hex.EncodeToString(encryptedKey)
	encryptedDataHex := hex.EncodeToString(ciphertext)

	fmt.Println("Encrypted data (use this for decryption):")
	fmt.Printf("%s:%s\n", encryptedKeyHex, encryptedDataHex)

	fmt.Println("\nDebug info:")
	fmt.Printf("Encrypted AES key (hex): %s\n", encryptedKeyHex)
	fmt.Printf("Encrypted data (hex): %s\n", encryptedDataHex)
}

func decryptMessage(privateKeyPath, input string) (map[string]interface{}, error) {
	// Load private key
	privateKey, err := crypto.LoadPrivateKey(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %v", err)
	}

	// Split the input into encrypted key and data
	parts := strings.Split(input, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid input format. Expected format: <encrypted_key_hex>:<encrypted_data_hex>")
	}

	encryptedKeyHex := parts[0]
	encryptedDataHex := parts[1]

	// Decode hex strings to bytes
	encryptedKey, err := hex.DecodeString(encryptedKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted key: %v", err)
	}

	encryptedData, err := hex.DecodeString(encryptedDataHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted data: %v", err)
	}

	// Decrypt the AES key with RSA
	aesKey, err := crypto.DecryptWithPrivateKey(privateKey, encryptedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt AES key: %v", err)
	}

	// Decrypt the data with AES
	jsonData, err := crypto.DecryptAES(aesKey, encryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %v", err)
	}

	// Unmarshal the JSON data
	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON data: %v", err)
	}

	return result, nil
}
