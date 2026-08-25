package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// 终端事件不带正文时，非流式聚合应使用累计的 output_text 增量重建 content。
func TestHandleChatBufferedStreamingResponse_RebuildsTextFromDeltas(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)

	sse := strings.Join([]string{
		`data: {"type":"response.output_text.delta","output_index":0,"delta":"Hello"}`,
		``,
		`data: {"type":"response.output_text.delta","output_index":0,"delta":", world"}`,
		``,
		`data: {"type":"response.completed","response":{"id":"resp_rebuild","status":"completed","output":[],"usage":{"input_tokens":10,"output_tokens":5}}}`,
		``,
		`data: [DONE]`,
		``,
	}, "\n")

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"x-request-id": []string{"rid-1"}},
		Body:       io.NopCloser(strings.NewReader(sse)),
	}

	svc := &OpenAIGatewayService{}
	result, err := svc.handleChatBufferedStreamingResponse(resp, c, "gpt-5.2", "gpt-5.2", time.Now())
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, 10, result.Usage.InputTokens)
	require.Equal(t, 5, result.Usage.OutputTokens)

	require.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		ID      string `json:"id"`
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "resp_rebuild", body.ID)
	require.Len(t, body.Choices, 1)
	require.Equal(t, "Hello, world", body.Choices[0].Message.Content)
	require.Equal(t, "stop", body.Choices[0].FinishReason)
}

// 终端事件自带正文时，不应与增量重复拼接。
func TestHandleChatBufferedStreamingResponse_NoDuplicateWhenTerminalHasText(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)

	sse := strings.Join([]string{
		`data: {"type":"response.output_text.delta","output_index":0,"delta":"full text"}`,
		``,
		`data: {"type":"response.completed","response":{"id":"resp_full","status":"completed","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"full text"}]}],"usage":{"input_tokens":3,"output_tokens":2}}}`,
		``,
	}, "\n")

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(sse)),
	}

	svc := &OpenAIGatewayService{}
	_, err := svc.handleChatBufferedStreamingResponse(resp, c, "gpt-5.2", "gpt-5.2", time.Now())
	require.NoError(t, err)

	var body struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Len(t, body.Choices, 1)
	require.Equal(t, "full text", body.Choices[0].Message.Content)
}

// OAuth 账号请求完全无法识别的模型名应返回 404 model_not_found，而不是兜底成 gpt-5.1。
func TestForwardAsChatCompletions_UnknownModelRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)

	account := &Account{
		ID:       1,
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
	}
	body := []byte(`{"model":"this-model-does-not-exist-999","messages":[{"role":"user","content":"hi"}]}`)

	svc := &OpenAIGatewayService{}
	result, err := svc.ForwardAsChatCompletions(context.Background(), c, account, body, "", "")
	require.Error(t, err)
	require.Nil(t, result)
	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Contains(t, rec.Body.String(), "model_not_found")
}
