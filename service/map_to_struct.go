package service

import (
	"encoding/json"
	"fmt"
)

// mapToStruct converts map[string]interface{} to struct
func mapToStruct(input any, target any) error {
	// map[string]interface{} -> json -> struct (target) is the easiest way to do this
	// it's not the most efficient way, but I currently expect like 15 requests per month so ¯\_(ツ)_/¯
	bytes, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("failed to marshal input: %w", err)
	}
	if err := json.Unmarshal(bytes, target); err != nil {
		return fmt.Errorf("failed to unmarshal input: %w", err)
	}
	return nil
}
