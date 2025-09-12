package handlers

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/tasiuskenways/scalable-ecommerce/crypto-service/internal/utils"
	"github.com/tasiuskenways/scalable-ecommerce/crypto-service/internal/utils/crypto"
)

func EncryptHandler(publicKeyPath string) fiber.Handler {
	return func(c *fiber.Ctx) error {

		// Only process JSON requests
		if c.Get("Content-Type") != "application/json" {
			return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid content type. Expected application/json")
		}

		var data interface{}
		if err := json.Unmarshal([]byte(c.Body()), &data); err != nil {
			return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request format. Expected JSON")
		}

		// Load public key
		publicKey, err := crypto.LoadPublicKey(publicKeyPath)
		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to load public key")
		}

		// Convert data to JSON
		jsonData, err := json.Marshal(data)
		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to marshal data to JSON")
		}

		// Generate a new AES key
		aesKey, err := crypto.GenerateAESKey()
		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate AES key")
		}

		// Encrypt the JSON data with AES
		ciphertext, err := crypto.EncryptAES(aesKey, jsonData)
		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to encrypt data")
		}

		// Encrypt the AES key with RSA
		encryptedKey, err := crypto.EncryptWithPublicKey(publicKey, aesKey)
		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to encrypt AES key")
		}

		// Output the encrypted key and data in a format that can be easily used for decryption
		encryptedKeyHex := hex.EncodeToString(encryptedKey)
		encryptedDataHex := hex.EncodeToString(ciphertext)

		encryptedData := fmt.Sprintf("%s:%s", encryptedKeyHex, encryptedDataHex)

		return utils.SuccessResponse(c, "Encrypted Successfully", encryptedData)
	}
}
