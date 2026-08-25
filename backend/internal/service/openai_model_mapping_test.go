package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveOpenAIForwardModel(t *testing.T) {
	tests := []struct {
		name               string
		account            *Account
		requestedModel     string
		defaultMappedModel string
		expectedModel      string
	}{
		{
			name: "falls back to group default when account has no mapping",
			account: &Account{
				Credentials: map[string]any{},
			},
			requestedModel:     "gpt-5.4",
			defaultMappedModel: "gpt-4o-mini",
			expectedModel:      "gpt-4o-mini",
		},
		{
			name: "preserves exact passthrough mapping instead of group default",
			account: &Account{
				Credentials: map[string]any{
					"model_mapping": map[string]any{
						"gpt-5.4": "gpt-5.4",
					},
				},
			},
			requestedModel:     "gpt-5.4",
			defaultMappedModel: "gpt-4o-mini",
			expectedModel:      "gpt-5.4",
		},
		{
			name: "preserves wildcard passthrough mapping instead of group default",
			account: &Account{
				Credentials: map[string]any{
					"model_mapping": map[string]any{
						"gpt-*": "gpt-5.4",
					},
				},
			},
			requestedModel:     "gpt-5.4",
			defaultMappedModel: "gpt-4o-mini",
			expectedModel:      "gpt-5.4",
		},
		{
			name: "uses account remap when explicit target differs",
			account: &Account{
				Credentials: map[string]any{
					"model_mapping": map[string]any{
						"gpt-5": "gpt-5.4",
					},
				},
			},
			requestedModel:     "gpt-5",
			defaultMappedModel: "gpt-4o-mini",
			expectedModel:      "gpt-5.4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveOpenAIForwardModel(tt.account, tt.requestedModel, tt.defaultMappedModel); got != tt.expectedModel {
				t.Fatalf("resolveOpenAIForwardModel(...) = %q, want %q", got, tt.expectedModel)
			}
		})
	}
}

func TestResolveOpenAIForwardModel_PreventsClaudeModelFromFallingBackToGpt51(t *testing.T) {
	account := &Account{
		Credentials: map[string]any{},
	}

	withoutDefault := resolveOpenAIForwardModel(account, "claude-opus-4-6", "")
	if got := normalizeCodexModel(withoutDefault); got != "gpt-5.1" {
		t.Fatalf("normalizeCodexModel(%q) = %q, want %q", withoutDefault, got, "gpt-5.1")
	}

	withDefault := resolveOpenAIForwardModel(account, "claude-opus-4-6", "gpt-5.4")
	if got := normalizeCodexModel(withDefault); got != "gpt-5.4" {
		t.Fatalf("normalizeCodexModel(%q) = %q, want %q", withDefault, got, "gpt-5.4")
	}
}

func TestNormalizeOpenAIModelAfterAccountMapping(t *testing.T) {
	tests := []struct {
		name                  string
		model                 string
		accountMappingMatched bool
		expected              string
	}{
		{
			name:                  "preserves explicitly mapped custom upstream model",
			model:                 "kimi-k3_b300_5",
			accountMappingMatched: true,
			expected:              "kimi-k3_b300_5",
		},
		{
			name:                  "preserves qualified custom upstream model",
			model:                 "vendor/kimi-k3_b300_5",
			accountMappingMatched: true,
			expected:              "vendor/kimi-k3_b300_5",
		},
		{
			name:                  "normalizes explicitly mapped known Codex alias",
			model:                 "gpt-5.1-codex-high",
			accountMappingMatched: true,
			expected:              "gpt-5.1-codex",
		},
		{
			name:                  "does not guess from custom model name containing gpt-5",
			model:                 "vendor-gpt-5-compatible",
			accountMappingMatched: true,
			expected:              "vendor-gpt-5-compatible",
		},
		{
			name:                  "keeps legacy fallback for unmatched custom request",
			model:                 "kimi-k3_b300_5",
			accountMappingMatched: false,
			expected:              "gpt-5.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeOpenAIModelAfterAccountMapping(tt.model, tt.accountMappingMatched); got != tt.expected {
				t.Fatalf("normalizeOpenAIModelAfterAccountMapping(%q, %v) = %q, want %q", tt.model, tt.accountMappingMatched, got, tt.expected)
			}
		})
	}
}

func TestResolveAndNormalizeExplicitCustomOpenAIModelMapping(t *testing.T) {
	account := &Account{Credentials: map[string]any{
		"model_mapping": map[string]any{
			"kimi-k3": "kimi-k3_b300_5",
		},
	}}

	mappedModel, matched := account.ResolveMappedModel("kimi-k3")
	require.True(t, matched)
	require.Equal(t, "kimi-k3_b300_5", mappedModel)
	require.Equal(t, "kimi-k3_b300_5", normalizeOpenAIModelAfterAccountMapping(mappedModel, matched))
}
