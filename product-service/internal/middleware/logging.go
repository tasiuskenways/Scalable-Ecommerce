package middleware

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

const maxDepth = 10

type LogEntry struct {
	Timestamp    string
	RequestID    string
	Method       string
	Path         string
	Query        string
	IP           string
	UserAgent    string
	Headers      map[string]string
	RequestBody  any
	StatusCode   int
	ResponseBody any
	Duration     int64
	Error        string
}

var logEntryPool = sync.Pool{
	New: func() any {
		return new(LogEntry)
	},
}

func RequestResponseLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		requestID := c.Locals("requestid").(string)

		entry := logEntryPool.Get().(*LogEntry)
		defer logEntryPool.Put(entry)
		*entry = LogEntry{}

		// Process request
		err := c.Next()

		duration := time.Since(start).Milliseconds()

		// Capture request body
		var requestBody any
		if len(c.Body()) > 0 && isJSONContent(c) {
			var body map[string]any
			if err := json.Unmarshal(c.Body(), &body); err == nil {
				requestBody = maskSensitiveData(body, 0)
			} else {
				requestBody = string(c.Body())
			}
		}

		// Capture response body
		var responseBody any
		respBody := c.Response().Body()
		if len(respBody) > 0 && isJSONResponse(c) {
			var body map[string]any
			if err := json.Unmarshal(respBody, &body); err == nil {
				responseBody = maskSensitiveData(body, 0)
			} else {
				responseBody = string(respBody)
			}
		}

		headers := make(map[string]string)
		c.Request().Header.VisitAll(func(key, value []byte) {
			k := string(key)
			if !isSensitiveHeader(k) {
				headers[k] = string(value)
			}
		})

		entry.Timestamp = start.Format(time.RFC3339)
		entry.RequestID = requestID
		entry.Method = c.Method()
		entry.Path = c.Path()
		entry.Query = string(c.Request().URI().QueryString())
		entry.IP = c.IP()
		entry.UserAgent = c.Get("User-Agent")
		entry.Headers = headers
		entry.RequestBody = requestBody
		entry.StatusCode = c.Response().StatusCode()
		entry.ResponseBody = responseBody
		entry.Duration = duration

		if err != nil {
			entry.Error = err.Error()
		}

		log.Info().
			Str("timestamp", entry.Timestamp).
			Str("request_id", entry.RequestID).
			Str("method", entry.Method).
			Str("path", entry.Path).
			Str("query", entry.Query).
			Str("ip", entry.IP).
			Str("user_agent", entry.UserAgent).
			Interface("headers", entry.Headers).
			Interface("request_body", entry.RequestBody).
			Int("status_code", entry.StatusCode).
			Interface("response_body", entry.ResponseBody).
			Int64("duration_ms", entry.Duration).
			Str("error", entry.Error).
			Send()

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

func maskSensitiveData(data map[string]any, depth int) map[string]any {
	if depth > maxDepth {
		return nil
	}

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
				masked[k] = maskSensitiveData(val, depth+1)
			default:
				masked[k] = v
			}
		}
	}

	return masked
}
