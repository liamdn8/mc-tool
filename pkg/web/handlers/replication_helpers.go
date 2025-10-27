package handlers

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

func (h *ReplicationHandler) getMCInternalAliases() ([]map[string]string, error) {
	cmd := exec.Command("mc", "alias", "list", "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	var aliases []map[string]string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var aliasData map[string]interface{}
		if err := json.Unmarshal([]byte(line), &aliasData); err != nil {
			continue
		}

		aliasName, okName := aliasData["alias"].(string)
		aliasURL, okURL := aliasData["URL"].(string)
		if !okName || !okURL {
			continue
		}

		aliases = append(aliases, map[string]string{
			"name": aliasName,
			"url":  aliasURL,
		})
	}

	return aliases, nil
}

func (h *ReplicationHandler) checkConsistency(data interface{}) bool {
	switch v := data.(type) {
	case map[string]string:
		if len(v) <= 1 {
			return true
		}

		var firstValue string
		first := true
		for _, value := range v {
			if first {
				firstValue = value
				first = false
				continue
			}

			if value != firstValue {
				return false
			}
		}
		return true
	case map[string]interface{}:
		if len(v) <= 1 {
			return true
		}

		var firstValue string
		first := true
		for _, value := range v {
			valueStr := fmt.Sprintf("%v", value)
			if first {
				firstValue = valueStr
				first = false
				continue
			}

			if valueStr != firstValue {
				return false
			}
		}
		return true
	}

	return true
}
