package services

import (
	"fmt"
	"time"

	"lio-ai/internal/models"
	"lio-ai/internal/repositories"
)

// UsageService handles business logic for usage tracking
type UsageService struct {
	usageRepo *repositories.UsageRepository
}

// NewUsageService creates a new usage service
func NewUsageService(usageRepo *repositories.UsageRepository) *UsageService {
	return &UsageService{
		usageRepo: usageRepo,
	}
}

// CalculateCost calculates the cost based on token usage and model
func (s *UsageService) CalculateCost(tokensInput, tokensOutput int, modelName string) (float64, error) {
	config, err := s.usageRepo.GetCostConfig(modelName)
	if err != nil {
		return 0, fmt.Errorf("failed to get cost config: %w", err)
	}

	// Calculate cost (prices are per 1000 tokens)
	inputCost := float64(tokensInput) * config.CostPerInputToken / 1000.0
	outputCost := float64(tokensOutput) * config.CostPerOutputToken / 1000.0
	totalCost := inputCost + outputCost

	return totalCost, nil
}

// TrackUsage tracks a usage event
func (s *UsageService) TrackUsage(req *models.UsageRequest) error {
	// Calculate cost
	cost, err := s.CalculateCost(req.TokensInput, req.TokensOutput, req.ModelUsed)
	if err != nil {
		return err
	}

	// Create usage metric
	metric := &models.UsageMetric{
		UserID:       req.UserID,
		RequestType:  req.RequestType,
		ResourceID:   req.ResourceID,
		TokensInput:  req.TokensInput,
		TokensOutput: req.TokensOutput,
		TokensTotal:  req.TokensInput + req.TokensOutput,
		ModelUsed:    req.ModelUsed,
		CostUSD:      cost,
		DurationMs:   req.DurationMs,
		Endpoint:     req.Endpoint,
		Success:      req.Success,
		ErrorMessage: req.ErrorMessage,
	}

	// Track the usage
	if err := s.usageRepo.TrackUsage(metric); err != nil {
		return fmt.Errorf("failed to track usage: %w", err)
	}

	// Update quota if successful
	if req.Success {
		if err := s.usageRepo.UpdateQuotaUsage(req.UserID, metric.TokensTotal, cost); err != nil {
			return fmt.Errorf("failed to update quota: %w", err)
		}
	}

	return nil
}

// CheckQuota checks if user has enough quota
func (s *UsageService) CheckQuota(userID string, tokensNeeded int, modelName string) (bool, error) {
	// Get or create user quota
	quota, err := s.usageRepo.GetUserQuota(userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user quota: %w", err)
	}

	// Check if daily/monthly reset is needed
	now := time.Now()
	if now.Sub(quota.LastResetDaily) >= 24*time.Hour {
		if err := s.usageRepo.ResetDailyQuota(userID); err != nil {
			return false, fmt.Errorf("failed to reset daily quota: %w", err)
		}
		quota.DailyTokensUsed = 0
		quota.DailyCostUsedUSD = 0.0
	}

	if now.Sub(quota.LastResetMonthly) >= 30*24*time.Hour {
		if err := s.usageRepo.ResetMonthlyQuota(userID); err != nil {
			return false, fmt.Errorf("failed to reset monthly quota: %w", err)
		}
		quota.MonthlyTokensUsed = 0
		quota.MonthlyCostUsedUSD = 0.0
	}

	// Check token limits
	if quota.DailyTokensUsed+tokensNeeded > quota.DailyTokenLimit {
		return false, nil
	}
	if quota.MonthlyTokensUsed+tokensNeeded > quota.MonthlyTokenLimit {
		return false, nil
	}

	// Estimate cost and check cost limits
	estimatedCost, err := s.CalculateCost(tokensNeeded/2, tokensNeeded/2, modelName)
	if err != nil {
		return false, fmt.Errorf("failed to calculate cost: %w", err)
	}

	if quota.DailyCostUsedUSD+estimatedCost > quota.DailyCostLimitUSD {
		return false, nil
	}
	if quota.MonthlyCostUsedUSD+estimatedCost > quota.MonthlyCostLimitUSD {
		return false, nil
	}

	return true, nil
}

// GetQuotaStatus retrieves the current quota status for a user
func (s *UsageService) GetQuotaStatus(userID string) (*models.QuotaStatus, error) {
	quota, err := s.usageRepo.GetUserQuota(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user quota: %w", err)
	}

	// Check if reset is needed
	now := time.Now()
	if now.Sub(quota.LastResetDaily) >= 24*time.Hour {
		if err := s.usageRepo.ResetDailyQuota(userID); err != nil {
			return nil, fmt.Errorf("failed to reset daily quota: %w", err)
		}
		quota.DailyTokensUsed = 0
		quota.DailyCostUsedUSD = 0.0
	}

	if now.Sub(quota.LastResetMonthly) >= 30*24*time.Hour {
		if err := s.usageRepo.ResetMonthlyQuota(userID); err != nil {
			return nil, fmt.Errorf("failed to reset monthly quota: %w", err)
		}
		quota.MonthlyTokensUsed = 0
		quota.MonthlyCostUsedUSD = 0.0
	}

	status := &models.QuotaStatus{
		UserID:              userID,
		DailyTokenLimit:     quota.DailyTokenLimit,
		DailyTokensUsed:     quota.DailyTokensUsed,
		DailyTokensRemaining: quota.DailyTokenLimit - quota.DailyTokensUsed,
		DailyTokensPercentUsed: float64(quota.DailyTokensUsed) / float64(quota.DailyTokenLimit) * 100,
		MonthlyTokenLimit:      quota.MonthlyTokenLimit,
		MonthlyTokensUsed:      quota.MonthlyTokensUsed,
		MonthlyTokensRemaining: quota.MonthlyTokenLimit - quota.MonthlyTokensUsed,
		MonthlyTokensPercentUsed: float64(quota.MonthlyTokensUsed) / float64(quota.MonthlyTokenLimit) * 100,
		DailyCostLimitUSD:        quota.DailyCostLimitUSD,
		DailyCostUsedUSD:         quota.DailyCostUsedUSD,
		DailyCostRemainingUSD:    quota.DailyCostLimitUSD - quota.DailyCostUsedUSD,
		DailyCostPercentUsed:     quota.DailyCostUsedUSD / quota.DailyCostLimitUSD * 100,
		MonthlyCostLimitUSD:      quota.MonthlyCostLimitUSD,
		MonthlyCostUsedUSD:       quota.MonthlyCostUsedUSD,
		MonthlyCostRemainingUSD:  quota.MonthlyCostLimitUSD - quota.MonthlyCostUsedUSD,
		MonthlyCostPercentUsed:   quota.MonthlyCostUsedUSD / quota.MonthlyCostLimitUSD * 100,
		LastResetDaily:           quota.LastResetDaily,
		LastResetMonthly:         quota.LastResetMonthly,
	}

	return status, nil
}

// GetUsageSummary retrieves aggregated usage for a user
func (s *UsageService) GetUsageSummary(userID, period string) (*models.UsageSummary, error) {
	summary, err := s.usageRepo.GetUsageSummary(userID, period)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage summary: %w", err)
	}

	// Get breakdown by endpoint
	endpoints, err := s.usageRepo.GetUsageByEndpoint(userID, period)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage by endpoint: %w", err)
	}

	summary.EndpointBreakdown = endpoints
	return summary, nil
}

// UpdateQuota updates the quota limits for a user
func (s *UsageService) UpdateQuota(userID string, req *models.QuotaUpdateRequest) error {
	updates := make(map[string]interface{})

	if req.DailyTokenLimit != nil {
		updates["daily_token_limit"] = *req.DailyTokenLimit
	}
	if req.MonthlyTokenLimit != nil {
		updates["monthly_token_limit"] = *req.MonthlyTokenLimit
	}
	if req.DailyCostLimitUSD != nil {
		updates["daily_cost_limit_usd"] = *req.DailyCostLimitUSD
	}
	if req.MonthlyCostLimitUSD != nil {
		updates["monthly_cost_limit_usd"] = *req.MonthlyCostLimitUSD
	}

	if len(updates) == 0 {
		return fmt.Errorf("no updates provided")
	}

	return s.usageRepo.UpdateUserQuota(userID, updates)
}
