package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

func ConvertToType(data interface{}, contentType string) (string, error) {
	switch contentType {
	case "application/json":
		parsedData := parseNumbers(data)
		jsonData, err := json.Marshal(parsedData)

		if err != nil {
			return "", fmt.Errorf("failed to marshal JSON: %w", err)
		}

		return string(jsonData), nil

	case "text/plain":
		return fmt.Sprint(data), nil

	default:
		return "", errors.New("unsupported content type")
	}
}

func parseNumbers(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			v[key] = parseNumbers(value)
		}
		return v
	case []interface{}:
		for i, value := range v {
			v[i] = parseNumbers(value)
		}
		return v
	case string:
		if num, err := strconv.ParseFloat(v, 64); err == nil {
			return num
		}
		return v
	default:
		return v
	}
}
