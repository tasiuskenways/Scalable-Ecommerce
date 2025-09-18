package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

type LogEntry struct {
	Timestamp    string            `json:"timestamp"`
	RequestID    string            `json:"request_id"`
	Method       string            `json:"method"`
	Path         string            `json:"path"`
	Query        string            `json:"query,omitempty"`
	IP           string            `json:"ip"`
	UserAgent    string            `json:"user_agent"`
	Headers      map[string]string `json:"headers,omitempty"`
	RequestBody  any               `json:"request_body,omitempty"`
	StatusCode   int               `json:"status_code"`
	ResponseBody any               `json:"response_body,omitempty"`
	Duration     string            `json:"duration"`
	Error        string            `json:"error,omitempty"`
}

// Object pool for LogEntry structs to reduce memory allocations
var logEntryPool = sync.Pool{
	New: func() any {
		return &LogEntry{}
	},
}

// Object pool for header maps
var headerMapPool = sync.Pool{
	New: func() any {
		return make(map[string]string, 10) // Pre-allocate for common header count
	},
}

const maxMaskingDepth = 5 // Limit recursion depth for performance

func RequestResponseLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Safe requestID extraction with fallback
		var requestID string
		if rid := c.Locals("requestid"); rid != nil {
			if ridStr, ok := rid.(string); ok {
				requestID = ridStr
			} else {
				requestID = "unknown"
			}
		} else {
			requestID = "missing"
		}

		// Get pooled objects
		logEntry := logEntryPool.Get().(*LogEntry)
		headers := headerMapPool.Get().(map[string]string)

		// Reset the maps
		for k := range headers {
			delete(headers, k)
		}

		// Lazy evaluation: only parse bodies if they're not too large
		var requestBody any
		bodyBytes := c.Body()
		if len(bodyBytes) > 0 && len(bodyBytes) < 10*1024 && isJSONContent(c) { // Max 10KB
			var body map[string]any
			if err := json.Unmarshal(bodyBytes, &body); err == nil {
				requestBody = maskSensitiveDataWithDepth(body, 0)
			}
		}

		// Capture request headers (excluding sensitive ones) - optimized
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

		// Lazy evaluation for response body
		var responseBody any
		respBody := c.Response().Body()
		if len(respBody) > 0 && len(respBody) < 10*1024 && isJSONResponse(c) { // Max 10KB
			var body map[string]any
			if err := json.Unmarshal(respBody, &body); err == nil {
				responseBody = maskSensitiveDataWithDepth(body, 0)
			}
		}

		// Populate log entry using pooled object
		logEntry.Timestamp = start.Format(time.RFC3339)
		logEntry.RequestID = requestID
		logEntry.Method = c.Method()
		logEntry.Path = c.Path()
		logEntry.Query = string(c.Request().URI().QueryString())
		logEntry.IP = c.IP()
		logEntry.UserAgent = c.Get("User-Agent")
		logEntry.Headers = headers
		logEntry.RequestBody = requestBody
		logEntry.StatusCode = c.Response().StatusCode()
		logEntry.ResponseBody = responseBody
		logEntry.Duration = formatDuration(duration)
		logEntry.Error = ""

		// Add error if exists
		if err != nil {
			logEntry.Error = err.Error()
		}

		// Use structured logging instead of fmt.Println
		if logJSON, marshalErr := json.Marshal(logEntry); marshalErr == nil {
			log.Println(string(logJSON))
		}

		// Return objects to pool
		logEntryPool.Put(logEntry)
		headerMapPool.Put(headers)

		return err
	}
}

// Optimized duration formatting to avoid string concatenation
func formatDuration(d time.Duration) string {
	ms := d.Nanoseconds() / 1000000
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	return fmt.Sprintf("%.2fs", float64(ms)/1000.0)
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

// Pre-compiled sensitive field set for faster lookup
var sensitiveFields = map[string]struct{}{
	"password":      {},
	"token":         {},
	"secret":        {},
	"api_key":       {},
	"apikey":        {},
	"access_token":  {},
	"refresh_token": {},
	"credit_card":   {},
	"card_number":   {},
	"cvv":           {},
	"ssn":           {},
}

// Optimized masking with depth limit to prevent stack overflow and improve performance
func maskSensitiveDataWithDepth(data map[string]any, depth int) map[string]any {
	if depth >= maxMaskingDepth {
		return map[string]any{"_truncated": "max_depth_reached"}
	}

	masked := make(map[string]any, len(data))
	for k, v := range data {
		keyLower := strings.ToLower(k)

		// Fast lookup using map instead of slice iteration
		isSensitive := false
		for field := range sensitiveFields {
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
				masked[k] = maskSensitiveDataWithDepth(val, depth+1)
			default:
				masked[k] = v
			}
		}
	}

	return masked
}