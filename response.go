package kra

import (
	"encoding/json"
	"strconv"
	"strings"
)

// APIResponse represents a normalized GavaConnect response.
type APIResponse struct {
	Data map[string]interface{}
	Meta ResponseMetadata
	Raw  map[string]interface{}
}

// ResponseMetadata captures response envelope information.
type ResponseMetadata struct {
	ResponseCode string
	ResponseDesc string
	Status       string
	ErrorCode    string
	ErrorMessage string
	RequestID    string
}

func normalizeAPIResponse(raw map[string]interface{}, statusCode int, endpoint string, body []byte) (*APIResponse, error) {
	meta := ResponseMetadata{
		ResponseCode: firstString(raw, "responseCode", "ResponseCode"),
		ResponseDesc: firstString(raw, "responseDesc", "ResponseDesc", "message", "Message"),
		Status:       firstString(raw, "status", "Status"),
		ErrorCode:    firstString(raw, "ErrorCode", "errorCode", "code"),
		ErrorMessage: firstString(raw, "ErrorMessage", "errorMessage"),
		RequestID:    firstString(raw, "requestId", "RequestId"),
	}

	if errMap, ok := raw["error"].(map[string]interface{}); ok {
		if meta.ErrorCode == "" {
			meta.ErrorCode = firstString(errMap, "code")
		}
		if meta.ErrorMessage == "" {
			meta.ErrorMessage = firstString(errMap, "message")
		}
	}

	data := extractPayload(raw)

	if isError(meta, raw) {
		msg := meta.ErrorMessage
		if msg == "" {
			msg = meta.ResponseDesc
		}
		if msg == "" {
			msg = "API request failed"
		}
		return nil, NewAPIError(statusCode, msg, endpoint, string(body))
	}

	return &APIResponse{
		Data: data,
		Meta: meta,
		Raw:  raw,
	}, nil
}

func extractPayload(raw map[string]interface{}) map[string]interface{} {
	if rd, ok := raw["responseData"].(map[string]interface{}); ok {
		return rd
	}
	if data, ok := raw["data"].(map[string]interface{}); ok {
		return data
	}
	// Legacy success structure
	if success, ok := raw["success"].(bool); ok && success {
		return raw["data"].(map[string]interface{})
	}
	return raw
}

func isError(meta ResponseMetadata, raw map[string]interface{}) bool {
	if meta.ErrorCode != "" {
		return true
	}
	if success, ok := raw["success"].(bool); ok {
		return !success
	}
	if meta.Status != "" && !strings.EqualFold(meta.Status, "ok") && !strings.EqualFold(meta.Status, "success") {
		return true
	}
	return false
}

func firstString(m map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if val, ok := m[key]; ok {
			switch v := val.(type) {
			case string:
				if strings.TrimSpace(v) != "" {
					return v
				}
			case json.Number:
				if s := v.String(); s != "" {
					return s
				}
			}
		}
	}
	return ""
}

func firstBool(m map[string]interface{}, keys ...string) (bool, bool) {
	for _, key := range keys {
		if val, ok := m[key]; ok {
			switch v := val.(type) {
			case bool:
				return v, true
			case string:
				s := strings.ToLower(strings.TrimSpace(v))
				if s == "true" || s == "1" || s == "yes" {
					return true, true
				}
				if s == "false" || s == "0" || s == "no" {
					return false, true
				}
			}
		}
	}
	return false, false
}

func firstFloat64(m map[string]interface{}, keys ...string) (float64, bool) {
	for _, key := range keys {
		if val, ok := m[key]; ok {
			switch v := val.(type) {
			case float64:
				return v, true
			case json.Number:
				f, err := v.Float64()
				if err == nil {
					return f, true
				}
			case string:
				f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
				if err == nil {
					return f, true
				}
			}
		}
	}
	return 0, false
}
