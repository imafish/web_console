package pb

import (
	"encoding/json"
	"log"
)

func (t *Task) AsJsonString() string {
	jsonData, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		log.Printf("Error marshalling Task to JSON: %v", err)
		return ""
	}
	return string(jsonData)
}
