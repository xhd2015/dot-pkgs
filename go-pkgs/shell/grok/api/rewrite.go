package api

import (
	"encoding/json"
	"strings"
)

func isGrokReasoningEffort(s string) bool {
	switch s {
	case "low", "medium", "high", "xhigh":
		return true
	}
	return false
}

// SplitModelEffort strips a trailing :effort suffix when it is a Grok
// reasoning level. "grok-4.6:high" → ("grok-4.6", "high").
func SplitModelEffort(model string) (base, effort string) {
	model = strings.TrimSpace(model)
	idx := strings.LastIndex(model, ":")
	if idx <= 0 {
		return model, ""
	}
	effort = model[idx+1:]
	if !isGrokReasoningEffort(effort) {
		return model, ""
	}
	return model[:idx], effort
}

// RewriteRequestJSON forces stream=true and moves a model:effort suffix into
// reasoning.effort when absent. Returns the unsuffixed model id.
func RewriteRequestJSON(body []byte) (model string, out []byte, err error) {
	if len(strings.TrimSpace(string(body))) == 0 {
		return "", body, nil
	}
	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return "", body, err
	}
	changed := false
	if raw, ok := data["model"].(string); ok {
		base, effort := SplitModelEffort(raw)
		model = base
		if effort != "" {
			data["model"] = base
			reasoning, _ := data["reasoning"].(map[string]any)
			if reasoning == nil {
				reasoning = map[string]any{}
			}
			if _, exists := reasoning["effort"]; !exists {
				reasoning["effort"] = effort
				data["reasoning"] = reasoning
			}
			changed = true
		}
	}
	if stream, ok := data["stream"].(bool); !ok || !stream {
		data["stream"] = true
		changed = true
	}
	if !changed {
		return model, body, nil
	}
	out, err = json.Marshal(data)
	if err != nil {
		return model, body, err
	}
	return model, out, nil
}
