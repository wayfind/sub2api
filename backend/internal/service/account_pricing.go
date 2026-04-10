package service

// ResolveModelPricing 解析模型定价（两级瀑布）。
//
// 优先级：
//  1. Account 人工确认价格（Extra["model_pricing"]）
//  2. LiteLLM 动态价格 + BillingService 硬编码回退（现有 GetModelPricing 链路）
//
// 返回 nil 表示全部未命中（调用方应按 cost=0 处理）。
// 当 Account 无 model_pricing 配置时，第 1 步返回 nil，直接走第 2 步，行为与改造前一致。
func ResolveModelPricing(account *Account, billingModel string, billingService *BillingService) *ModelPricing {
	if billingService == nil || billingModel == "" {
		return nil
	}

	// 1. Account 人工确认价格
	if account != nil {
		if pricing := account.GetModelPricingOverride(billingModel); pricing != nil {
			pricing.Source = "account"
			return pricing
		}
	}

	// 2. LiteLLM + 硬编码回退（现有链路，GetModelPricing 内部已包含 fallback）
	if pricing, err := billingService.GetModelPricing(billingModel); err == nil && pricing != nil {
		// Source 已在 GetModelPricing 内部设置（"litellm" 或 "fallback"）
		return pricing
	}

	return nil
}
