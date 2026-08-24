package service

import "strings"

// resolveOpenAIForwardModel determines the upstream model for OpenAI-compatible
// forwarding. Group-level default mapping only applies when the account itself
// did not match any explicit model_mapping rule.
func resolveOpenAIForwardModel(account *Account, requestedModel, defaultMappedModel string) string {
	if account == nil {
		if defaultMappedModel != "" {
			return defaultMappedModel
		}
		return requestedModel
	}

	mappedModel, matched := account.ResolveMappedModel(requestedModel)
	if !matched && defaultMappedModel != "" {
		return defaultMappedModel
	}
	return mappedModel
}

// normalizeOpenAIModelAfterAccountMapping keeps unknown explicit mapping targets
// authoritative. Targets recognized by the built-in Codex alias table retain
// the legacy normalization behavior.
func normalizeOpenAIModelAfterAccountMapping(model string, accountMappingMatched bool) string {
	if accountMappingMatched {
		modelID := strings.TrimSpace(model)
		if idx := strings.LastIndex(modelID, "/"); idx >= 0 {
			modelID = modelID[idx+1:]
		}
		if getNormalizedCodexModel(modelID) == "" {
			return model
		}
	}
	return normalizeCodexModel(model)
}
