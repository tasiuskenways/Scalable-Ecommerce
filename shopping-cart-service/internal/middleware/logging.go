package middleware

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type LogEntry struct {
	Timestamp    string                 `json:"timestamp"`
	RequestID    string                 `json:"request_id"`
	Method       string                 `json:"method"`
	Path         string                 `json:"path"`
	Query        string                 `json:"query,omitempty"`
	IP           string                 `json:"ip"`
	UserAgent    string                 `json:"user_agent"`
	Headers      map[string]string      `json:"headers,omitempty"`
	RequestBody  any                    `json:"request_body,omitempty"`
	StatusCode   int                    `json:"status_code"`
	ResponseBody any                    `json:"response_body,omitempty"`
	Duration     string                 `json:"duration"`
	Error        string                 `json:"error,omitempty"`
}

func RequestResponseLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		requestID := c.Locals("requestid").(string)

		// Capture request body
		var requestBody any
		if len(c.Body()) > 0 && isJSONContent(c) {
			var body map[string]any
			if err := json.Unmarshal(c.Body(), &body); err == nil {
				// Mask sensitive fields
				requestBody = maskSensitiveData(body)
			} else {
				requestBody = string(c.Body())
			}
		}

		// Capture request headers (excluding sensitive ones)
		headers := make(map[string]string)
		c.Request().Header.VisitAll(func(key, value []byte) {
			k := string(key)
			if !isSensitiveHeader(k) {
				headers[k] = string(value)
			}
		})

		// Process request
		err := c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Capture response body (Fiber stores it internally)
		var responseBody any
		respBody := c.Response().Body()
		if len(respBody) > 0 && isJSONResponse(c) {
			var body map[string]any
			if err := json.Unmarshal(respBody, &body); err == nil {
				responseBody = maskSensitiveData(body)
			} else {
				responseBody = string(respBody)
			}
		}

		// Create log entry
		logEntry := LogEntry{
			Timestamp:    start.Format(time.RFC3339),
			RequestID:    requestID,
			Method:       c.Method(),
			Path:         c.Path(),
			Query:        string(c.Request().URI().QueryString()),
			IP:           c.IP(),
			UserAgent:    c.Get("User-Agent"),
			Headers:      headers,
			RequestBody:  requestBody,
			StatusCode:   c.Response().StatusCode(),
			ResponseBody: responseBody,
			Duration:     fmt.Sprintf("%dms", duration.Milliseconds()),
		}

		// Add error if exists
		if err != nil {
			logEntry.Error = err.Error()
		}

		// Log the entry
		logJSON, _ := json.Marshal(logEntry)
		fmt.Println(string(logJSON))

		return err
	}
}

func isJSONContent(c *fiber.Ctx) bool {
	contentType := c.Get("Content-Type")
	return strings.Contains(contentType, "application/json")
}

func isJSONResponse(c *fiber.Ctx) bool {
	contentType := c.GetRespHeader("Content-Type")
	return strings.Contains(contentType, "application/json")
}

func isSensitiveHeader(header string) bool {
	sensitive := []string{
		"Authorization",
		"Cookie",
		"Set-Cookie",
		"X-Auth-Token",
		"X-Api-Key",
	}

	headerLower := strings.ToLower(header)
	for _, s := range sensitive {
		if strings.ToLower(s) == headerLower {
			return true
		}
	}
	return false
}

func maskSensitiveData(data map[string]any) map[string]any {
	sensitiveFields := []string{
		"password",
		"token",
		"secret",
		"api_key",
		"apikey",
		"access_token",
		"refresh_token",
		"credit_card",
		"card_number",
		"cvv",
		"ssn",
	}

	masked := make(map[string]any)
	for k, v := range data {
		keyLower := strings.ToLower(k)
		isSensitive := false

		for _, field := range sensitiveFields {
			if strings.Contains(keyLower, field) {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			masked[k] = "***MASKED***"
		} else {
			switch val := v.(type) {
			case map[string]any:
				masked[k] = maskSensitiveData(val)
			default:
				masked[k] = v
			}
		}
	}

	return masked
}