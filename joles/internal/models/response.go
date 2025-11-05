package models

// APIResponse represents a standardized API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// APIError represents an error in API response
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Meta represents metadata for paginated responses
type Meta struct {
	Page       int `json:"page,omitempty"`
	PageSize   int `json:"page_size,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
	TotalCount int `json:"total_count,omitempty"`
}

// PaginationRequest represents pagination parameters
type PaginationRequest struct {
	Page     int `form:"page" json:"page"`
	PageSize int `form:"page_size" json:"page_size"`
}

// GetPagination returns validated pagination values
func (p *PaginationRequest) GetPagination() (int, int) {
	page := p.Page
	if page < 1 {
		page = 1
	}

	pageSize := p.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return page, pageSize
}

// GetOffset calculates the offset for database queries
func (p *PaginationRequest) GetOffset() int {
	page, pageSize := p.GetPagination()
	return (page - 1) * pageSize
}

// GetLimit returns the limit for database queries
func (p *PaginationRequest) GetLimit() int {
	_, pageSize := p.GetPagination()
	return pageSize
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Gateway   string            `json:"gateway"`
	Backend   string            `json:"backend"`
	Database  string            `json:"database"`
	Timestamp string            `json:"timestamp"`
	Version   string            `json:"version"`
	Uptime    string            `json:"uptime,omitempty"`
	Checks    map[string]string `json:"checks,omitempty"`
}

// MetricsResponse represents system metrics
type MetricsResponse struct {
	RequestsTotal      int64              `json:"requests_total"`
	RequestsSuccessful int64              `json:"requests_successful"`
	RequestsFailed     int64              `json:"requests_failed"`
	AverageLatencyMs   float64            `json:"average_latency_ms"`
	ActiveUsers        int                `json:"active_users"`
	TotalUsers         int                `json:"total_users"`
	TotalChats         int                `json:"total_chats"`
	TotalDocuments     int                `json:"total_documents"`
	TotalTokensUsed    int                `json:"total_tokens_used"`
	TotalCostUSD       float64            `json:"total_cost_usd"`
	EndpointStats      []EndpointStat     `json:"endpoint_stats,omitempty"`
	ModelStats         []ModelStat        `json:"model_stats,omitempty"`
}

// EndpointStat represents statistics for an endpoint
type EndpointStat struct {
	Endpoint      string  `json:"endpoint"`
	RequestCount  int     `json:"request_count"`
	AverageTimeMs float64 `json:"average_time_ms"`
	ErrorRate     float64 `json:"error_rate"`
}

// ModelStat represents statistics for a model
type ModelStat struct {
	ModelName    string  `json:"model_name"`
	RequestCount int     `json:"request_count"`
	TotalTokens  int     `json:"total_tokens"`
	TotalCostUSD float64 `json:"total_cost_usd"`
}

// ErrorCode constants
const (
	ErrCodeValidation     = "VALIDATION_ERROR"
	ErrCodeNotFound       = "NOT_FOUND"
	ErrCodeUnauthorized   = "UNAUTHORIZED"
	ErrCodeForbidden      = "FORBIDDEN"
	ErrCodeQuotaExceeded  = "QUOTA_EXCEEDED"
	ErrCodeRateLimited    = "RATE_LIMITED"
	ErrCodeInternal       = "INTERNAL_ERROR"
	ErrCodeBadRequest     = "BAD_REQUEST"
	ErrCodeServiceDown    = "SERVICE_DOWN"
)
