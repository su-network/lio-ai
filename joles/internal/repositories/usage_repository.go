package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"lio-ai/internal/models"
)

// UsageRepository handles database operations for usage tracking
type UsageRepository struct {
	db *sql.DB
}

// NewUsageRepository creates a new usage repository
func NewUsageRepository(db *sql.DB) *UsageRepository {
	return &UsageRepository{db: db}
}

// TrackUsage records a usage metric
func (r *UsageRepository) TrackUsage(metric *models.UsageMetric) error {
	query := `
		INSERT INTO usage_metrics (
			user_id, request_type, resource_id, tokens_input, tokens_output,
			tokens_total, model_used, cost_usd, duration_ms, endpoint,
			success, error_message, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	result, err := r.db.Exec(query,
		metric.UserID, metric.RequestType, metric.ResourceID,
		metric.TokensInput, metric.TokensOutput, metric.TokensTotal,
		metric.ModelUsed, metric.CostUSD, metric.DurationMs,
		metric.Endpoint, metric.Success, metric.ErrorMessage, now,
	)
	if err != nil {
		return fmt.Errorf("failed to track usage: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	metric.ID = id
	metric.CreatedAt = now
	return nil
}

// GetUserQuota retrieves or creates a user quota
func (r *UsageRepository) GetUserQuota(userID string) (*models.UserQuota, error) {
	query := `
		SELECT id, user_id, daily_token_limit, monthly_token_limit,
			daily_tokens_used, monthly_tokens_used, daily_cost_limit_usd,
			monthly_cost_limit_usd, daily_cost_used_usd, monthly_cost_used_usd,
			last_reset_daily, last_reset_monthly, created_at, updated_at
		FROM user_quotas
		WHERE user_id = ?
	`

	quota := &models.UserQuota{}
	err := r.db.QueryRow(query, userID).Scan(
		&quota.ID, &quota.UserID, &quota.DailyTokenLimit, &quota.MonthlyTokenLimit,
		&quota.DailyTokensUsed, &quota.MonthlyTokensUsed, &quota.DailyCostLimitUSD,
		&quota.MonthlyCostLimitUSD, &quota.DailyCostUsedUSD, &quota.MonthlyCostUsedUSD,
		&quota.LastResetDaily, &quota.LastResetMonthly, &quota.CreatedAt, &quota.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Create default quota
		return r.CreateUserQuota(userID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user quota: %w", err)
	}

	return quota, nil
}

// CreateUserQuota creates a new user quota with defaults
func (r *UsageRepository) CreateUserQuota(userID string) (*models.UserQuota, error) {
	query := `
		INSERT INTO user_quotas (user_id, created_at, updated_at)
		VALUES (?, ?, ?)
	`

	now := time.Now()
	result, err := r.db.Exec(query, userID, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create user quota: %w", err)
	}

	id, _ := result.LastInsertId()
	
	return &models.UserQuota{
		ID:                  id,
		UserID:              userID,
		DailyTokenLimit:     100000,
		MonthlyTokenLimit:   3000000,
		DailyTokensUsed:     0,
		MonthlyTokensUsed:   0,
		DailyCostLimitUSD:   10.0,
		MonthlyCostLimitUSD: 300.0,
		DailyCostUsedUSD:    0.0,
		MonthlyCostUsedUSD:  0.0,
		LastResetDaily:      now,
		LastResetMonthly:    now,
		CreatedAt:           now,
		UpdatedAt:           now,
	}, nil
}

// UpdateQuotaUsage updates the quota usage
func (r *UsageRepository) UpdateQuotaUsage(userID string, tokens int, cost float64) error {
	query := `
		UPDATE user_quotas
		SET daily_tokens_used = daily_tokens_used + ?,
			monthly_tokens_used = monthly_tokens_used + ?,
			daily_cost_used_usd = daily_cost_used_usd + ?,
			monthly_cost_used_usd = monthly_cost_used_usd + ?,
			updated_at = ?
		WHERE user_id = ?
	`

	_, err := r.db.Exec(query, tokens, tokens, cost, cost, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update quota usage: %w", err)
	}

	return nil
}

// ResetDailyQuota resets daily usage if needed
func (r *UsageRepository) ResetDailyQuota(userID string) error {
	query := `
		UPDATE user_quotas
		SET daily_tokens_used = 0,
			daily_cost_used_usd = 0.0,
			last_reset_daily = ?,
			updated_at = ?
		WHERE user_id = ?
	`

	now := time.Now()
	_, err := r.db.Exec(query, now, now, userID)
	return err
}

// ResetMonthlyQuota resets monthly usage if needed
func (r *UsageRepository) ResetMonthlyQuota(userID string) error {
	query := `
		UPDATE user_quotas
		SET monthly_tokens_used = 0,
			monthly_cost_used_usd = 0.0,
			last_reset_monthly = ?,
			updated_at = ?
		WHERE user_id = ?
	`

	now := time.Now()
	_, err := r.db.Exec(query, now, now, userID)
	return err
}

// GetCostConfig retrieves cost configuration for a model
func (r *UsageRepository) GetCostConfig(modelName string) (*models.CostConfig, error) {
	query := `
		SELECT id, model_name, cost_per_input_token, cost_per_output_token,
			operation_type, is_active, created_at, updated_at
		FROM cost_config
		WHERE model_name = ? AND is_active = 1
	`

	config := &models.CostConfig{}
	err := r.db.QueryRow(query, modelName).Scan(
		&config.ID, &config.ModelName, &config.CostPerInputToken,
		&config.CostPerOutputToken, &config.OperationType, &config.IsActive,
		&config.CreatedAt, &config.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Return default config
		return r.GetCostConfig("default")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get cost config: %w", err)
	}

	return config, nil
}

// GetUsageSummary retrieves aggregated usage for a user
func (r *UsageRepository) GetUsageSummary(userID, period string) (*models.UsageSummary, error) {
	var whereClause string
	now := time.Now()

	switch period {
	case "daily":
		whereClause = fmt.Sprintf("AND created_at >= '%s'", now.AddDate(0, 0, -1).Format(time.RFC3339))
	case "monthly":
		whereClause = fmt.Sprintf("AND created_at >= '%s'", now.AddDate(0, -1, 0).Format(time.RFC3339))
	default:
		whereClause = ""
	}

	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total_requests,
			SUM(CASE WHEN success = 1 THEN 1 ELSE 0 END) as successful_requests,
			SUM(CASE WHEN success = 0 THEN 1 ELSE 0 END) as failed_requests,
			COALESCE(SUM(tokens_input), 0) as total_tokens_input,
			COALESCE(SUM(tokens_output), 0) as total_tokens_output,
			COALESCE(SUM(tokens_total), 0) as total_tokens,
			COALESCE(SUM(cost_usd), 0.0) as total_cost_usd,
			COALESCE(AVG(duration_ms), 0) as average_duration_ms,
			SUM(CASE WHEN request_type = 'chat' THEN 1 ELSE 0 END) as chat_requests,
			SUM(CASE WHEN request_type = 'code_generation' THEN 1 ELSE 0 END) as code_gen_requests
		FROM usage_metrics
		WHERE user_id = ? %s
	`, whereClause)

	summary := &models.UsageSummary{
		UserID: userID,
		Period: period,
		ModelsUsed: make(map[string]int),
	}

	err := r.db.QueryRow(query, userID).Scan(
		&summary.TotalRequests, &summary.SuccessfulRequests, &summary.FailedRequests,
		&summary.TotalTokensInput, &summary.TotalTokensOutput, &summary.TotalTokens,
		&summary.TotalCostUSD, &summary.AverageDurationMs, &summary.ChatRequests,
		&summary.CodeGenRequests,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get usage summary: %w", err)
	}

	return summary, nil
}

// GetUsageByEndpoint retrieves usage breakdown by endpoint
func (r *UsageRepository) GetUsageByEndpoint(userID, period string) ([]models.UsageByEndpoint, error) {
	var whereClause string
	now := time.Now()

	switch period {
	case "daily":
		whereClause = fmt.Sprintf("AND created_at >= '%s'", now.AddDate(0, 0, -1).Format(time.RFC3339))
	case "monthly":
		whereClause = fmt.Sprintf("AND created_at >= '%s'", now.AddDate(0, -1, 0).Format(time.RFC3339))
	default:
		whereClause = ""
	}

	query := fmt.Sprintf(`
		SELECT 
			endpoint,
			COUNT(*) as request_count,
			COALESCE(SUM(tokens_total), 0) as total_tokens,
			COALESCE(SUM(cost_usd), 0.0) as total_cost_usd,
			COALESCE(AVG(duration_ms), 0) as average_duration_ms,
			CAST(SUM(CASE WHEN success = 1 THEN 1 ELSE 0 END) AS REAL) / COUNT(*) * 100 as success_rate
		FROM usage_metrics
		WHERE user_id = ? %s
		GROUP BY endpoint
		ORDER BY request_count DESC
	`, whereClause)

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage by endpoint: %w", err)
	}
	defer rows.Close()

	var results []models.UsageByEndpoint
	for rows.Next() {
		var usage models.UsageByEndpoint
		err := rows.Scan(
			&usage.Endpoint, &usage.RequestCount, &usage.TotalTokens,
			&usage.TotalCostUSD, &usage.AverageDurationMs, &usage.SuccessRate,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage by endpoint: %w", err)
		}
		results = append(results, usage)
	}

	return results, nil
}

// UpdateUserQuota updates quota limits
func (r *UsageRepository) UpdateUserQuota(userID string, updates map[string]interface{}) error {
	query := `
		UPDATE user_quotas
		SET daily_token_limit = COALESCE(?, daily_token_limit),
			monthly_token_limit = COALESCE(?, monthly_token_limit),
			daily_cost_limit_usd = COALESCE(?, daily_cost_limit_usd),
			monthly_cost_limit_usd = COALESCE(?, monthly_cost_limit_usd),
			updated_at = ?
		WHERE user_id = ?
	`

	_, err := r.db.Exec(query,
		updates["daily_token_limit"],
		updates["monthly_token_limit"],
		updates["daily_cost_limit_usd"],
		updates["monthly_cost_limit_usd"],
		time.Now(),
		userID,
	)

	return err
}
