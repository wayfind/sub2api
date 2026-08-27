package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestIsOpaqueUpstreamReject 用真实上游响应体做判定回归。
// 样本取自生产实测：金山云 ELB 的裸文本 400 必须判为前置层拒绝，
// 而模型 API / 中转层给出的结构化 400 必须保持原语义（不放行 failover）。
func TestIsOpaqueUpstreamReject(t *testing.T) {
	opaque := map[string]string{
		"金山云 ELB 裸文本":     "Bad Request",
		"带尾部换行的裸文本":       "Bad Request\n",
		"空 body":          "",
		"仅空白":             "   \r\n ",
		"nginx HTML 错误页":  `<html><head><title>400 Bad Request</title></head></html>`,
		"截断的 JSON":        `{"error":{"type":"Bad`,
	}
	for name, body := range opaque {
		t.Run("opaque/"+name, func(t *testing.T) {
			require.True(t, isOpaqueUpstreamReject([]byte(body)))
		})
	}

	structured := map[string]string{
		"KSYUN 结构化 400":   `{"error":{"type":"Bad Request","message":"Invalid JSON data: Failed to deserialize the JSON body into the target type"},"type":"error"}`,
		"Anthropic 风格":    `{"type":"error","error":{"type":"invalid_request_error","message":"max_tokens is too large"}}`,
		"sglang-proxy 风格": `{"type":"error","error":{"type":"invalid_request_error","message":"invalid JSON: unexpected end of JSON input"}}`,
		"OpenAI 风格":       `{"error":{"message":"Invalid request","type":"invalid_request_error","code":null}}`,
		"带前后空白的 JSON":     "\n  {\"error\":{\"message\":\"x\"}}  \n",
	}
	for name, body := range structured {
		t.Run("structured/"+name, func(t *testing.T) {
			require.False(t, isOpaqueUpstreamReject([]byte(body)))
		})
	}
}
