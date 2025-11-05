package models

import "time"

// UsageMetric represents detailed usage tracking for a user
type UsageMetric struct {
	ID              int64     `json:"id"`
	UserID          string    `json:"user_id"`
	RequestType     string    `json:"request_type"` // "chat", "code_generation"
	ResourceID      int64     `json:"resource_id,omitempty"` // ChatID, DocumentID, etc.
	TokensInput     int       `json:"tokens_input"`
	TokensOutput    int       `json:"tokens_output"`
	TokensTotal     int       `json:"tokens_total"`
	ModelUsed       string    `json:"model_used"`
	CostUSD         float64   `json:"cost_usd"`
	DurationMs      int64     `json:"duration_ms"`
	Endpoint        string    `json:"endpoint"`
	Success         bool      `json:"success"`
	ErrorMessage    string    `json:"error_message,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// UserQuota represents usage limits and quotas for a user
type UserQuota struct {
	ID                  int64     `json:"id"`
	UserID              string    `json:"user_id"`
	DailyTokenLimit     int       `json:"daily_token_limit"`
	MonthlyTokenLimit   int       `json:"monthly_token_limit"`
	DailyTokensUsed     int       `json:"daily_tokens_used"`
	MonthlyTokensUsed   int       `json:"monthly_tokens_used"`
	DailyCostLimitUSD   float64   `json:"daily_cost_limit_usd"`
	MonthlyCostLimitUSD float64   `json:"monthly_cost_limit_usd"`
	DailyCostUsedUSD    float64   `json:"daily_cost_used_usd"`
	MonthlyCostUsedUSD  float64   `json:"monthly_cost_used_usd"`
	LastResetDaily      time.Time `json:"last_reset_daily"`
	LastResetMonthly    time.Time `json:"last_reset_monthly"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// CostConfig represents pricing configuration for different models and operations
type CostConfig struct {
	ID              int64     `json:"id"`
	ModelName       string    `json:"model_name"`
	CostPerInputToken  float64   `json:"cost_per_input_token"`  // USD per token
	CostPerOutputToken float64   `json:"cost_per_output_token"` // USD per token
	OperationType   string    `json:"operation_type"` // "chat", "code_generation", "embedding"
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// UsageSummary represents aggregated usage statistics
type UsageSummary struct {
	UserID              string              `json:"user_id"`
	Period              string              `json:"period"` // "daily", "monthly", "all_time"
	TotalRequests       int                 `json:"total_requests"`
	SuccessfulRequests  int                 `json:"successful_requests"`
	FailedRequests      int                 `json:"failed_requests"`
	TotalTokensInput    int                 `json:"total_tokens_input"`
	TotalTokensOutput   int                 `json:"total_tokens_output"`
	TotalTokens         int                 `json:"total_tokens"`
	TotalCostUSD        float64             `json:"total_cost_usd"`
	AverageDurationMs   float64             `json:"average_duration_ms"`
	ChatRequests        int                 `json:"chat_requests"`
	CodeGenRequests     int                 `json:"code_gen_requests"`
	ModelsUsed          map[string]int      `json:"models_used"`
	EndpointBreakdown   []UsageByEndpoint   `json:"endpoint_breakdown"`
}

// UsageByEndpoint represents usage breakdown by API endpoint
type UsageByEndpoint struct {
	Endpoint          string  `json:"endpoint"`
	RequestCount      int     `json:"request_count"`
	TotalTokens       int     `json:"total_tokens"`
	TotalCostUSD      float64 `json:"total_cost_usd"`
	AverageDurationMs float64 `json:"average_duration_ms"`
	SuccessRate       float64 `json:"success_rate"`
}

// QuotaStatus represents current quota usage status
type QuotaStatus struct {
	UserID                   string    `json:"user_id"`
	DailyTokensUsed          int       `json:"daily_tokens_used"`
	DailyTokenLimit          int       `json:"daily_token_limit"`
	DailyTokensRemaining     int       `json:"daily_tokens_remaining"`
	DailyTokensPercentUsed   float64   `json:"daily_tokens_percent_used"`
	MonthlyTokensUsed        int       `json:"monthly_tokens_used"`
	MonthlyTokenLimit        int       `json:"monthly_token_limit"`
	MonthlyTokensRemaining   int       `json:"monthly_tokens_remaining"`
	MonthlyTokensPercentUsed float64   `json:"monthly_tokens_percent_used"`
	DailyCostUsedUSD         float64   `json:"daily_cost_used_usd"`
	DailyCostLimitUSD        float64   `json:"daily_cost_limit_usd"`
	DailyCostRemainingUSD    float64   `json:"daily_cost_remaining_usd"`
	DailyCostPercentUsed     float64   `json:"daily_cost_percent_used"`
	MonthlyCostUsedUSD       float64   `json:"monthly_cost_used_usd"`
	MonthlyCostLimitUSD      float64   `json:"monthly_cost_limit_usd"`
	MonthlyCostRemainingUSD  float64   `json:"monthly_cost_remaining_usd"`
	MonthlyCostPercentUsed   float64   `json:"monthly_cost_percent_used"`
	LastResetDaily           time.Time `json:"last_reset_daily"`
	LastResetMonthly         time.Time `json:"last_reset_monthly"`
}

// UsageRequest represents a request to track usage
type UsageRequest struct {
	UserID       string `json:"user_id" binding:"required"`
	RequestType  string `json:"request_type" binding:"required"`
	ResourceID   int64  `json:"resource_id,omitempty"`
	TokensInput  int    `json:"tokens_input"`
	TokensOutput int    `json:"tokens_output"`
	ModelUsed    string `json:"model_used"`
	Endpoint     string `json:"endpoint"`
	DurationMs   int64  `json:"duration_ms"`
	Success      bool   `json:"success"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// QuotaUpdateRequest represents a request to update user quota
type QuotaUpdateRequest struct {
	DailyTokenLimit     *int     `json:"daily_token_limit,omitempty"`
	MonthlyTokenLimit   *int     `json:"monthly_token_limit,omitempty"`
	DailyCostLimitUSD   *float64 `json:"daily_cost_limit_usd,omitempty"`
	MonthlyCostLimitUSD *float64 `json:"monthly_cost_limit_usd,omitempty"`
}
