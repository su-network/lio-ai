"""Data models for code generation requests and responses"""
from pydantic import BaseModel, Field
from typing import List, Optional, Dict, Any
from datetime import datetime
from enum import Enum


class ComplexityLevel(str, Enum):
    """Code complexity levels"""
    SIMPLE = "simple"
    INTERMEDIATE = "intermediate"
    ADVANCED = "advanced"


class ModelProvider(str, Enum):
    """Supported model providers"""
    OPENAI = "openai"
    ANTHROPIC = "anthropic"
    GOOGLE = "google"
    COHERE = "cohere"
    OLLAMA = "ollama"
    LOCAL = "local"


class CodeGenRequest(BaseModel):
    """Code generation request schema"""
    request_id: str = Field(..., description="Unique request identifier")
    user_id: str = Field(..., description="User identifier")
    language: str = Field(..., description="Programming language (e.g., python, go, javascript)")
    framework: Optional[str] = Field(None, description="Framework/library (e.g., fastapi, gin, express)")
    complexity: ComplexityLevel = Field(..., description="Code complexity level")
    prompt: str = Field(..., min_length=10, description="Code generation prompt")
    context: Optional[str] = Field(None, description="Additional context or existing code")
    constraints: List[str] = Field(default_factory=list, description="Code constraints")
    selected_models: List[str] = Field(..., min_length=1, description="Array of model IDs to use")
    max_models: Optional[int] = Field(None, ge=1, le=10, description="Max models to auto-select")
    timeout: int = Field(30, ge=5, le=300, description="Request timeout in seconds")
    include_tests: bool = Field(False, description="Generate unit tests")
    include_docs: bool = Field(True, description="Generate documentation")
    stream_response: bool = Field(False, description="Enable streaming responses")
    temperature: Optional[float] = Field(None, ge=0.0, le=2.0, description="Model temperature")
    max_tokens: Optional[int] = Field(None, ge=100, le=8000, description="Max output tokens")


class QualityMetrics(BaseModel):
    """Code quality metrics"""
    cyclomatic_complexity: int = Field(0, description="Cyclomatic complexity score")
    lines_of_code: int = Field(0, description="Total lines of code")
    readability_score: float = Field(0.0, ge=0, le=100, description="Readability score (0-100)")
    security_score: float = Field(0.0, ge=0, le=100, description="Security score (0-100)")
    test_coverage: Optional[float] = Field(None, ge=0, le=100, description="Test coverage percentage")
    documentation_score: float = Field(0.0, ge=0, le=100, description="Documentation quality (0-100)")


class ModelResponse(BaseModel):
    """Response from a single model"""
    model_id: str = Field(..., description="Model identifier")
    model_name: str = Field(..., description="Human-readable model name")
    provider: str = Field(..., description="Model provider")
    generated_code: str = Field(..., description="Generated code")
    confidence_score: float = Field(..., ge=0, le=1, description="Model confidence (0-1)")
    execution_time_ms: float = Field(..., description="Execution time in milliseconds")
    tokens_used: int = Field(0, description="Total tokens consumed")
    prompt_tokens: int = Field(0, description="Input tokens")
    completion_tokens: int = Field(0, description="Output tokens")
    quality_metrics: QualityMetrics = Field(default_factory=QualityMetrics)
    syntax_valid: bool = Field(False, description="Syntax validation result")
    warnings: List[str] = Field(default_factory=list, description="Warnings or suggestions")
    error: Optional[str] = Field(None, description="Error message if failed")


class CodeGenResponse(BaseModel):
    """Complete code generation response"""
    request_id: str = Field(..., description="Original request ID")
    status: str = Field(..., description="Response status (success, partial, failed)")
    total_time_ms: float = Field(..., description="Total processing time")
    model_responses: List[ModelResponse] = Field(..., description="Individual model responses")
    consensus_code: Optional[str] = Field(None, description="Consensus/merged code")
    best_model: Optional[str] = Field(None, description="Recommended best model ID")
    timestamp: datetime = Field(default_factory=datetime.utcnow)
    metadata: Dict[str, Any] = Field(default_factory=dict)


class ModelCapability(BaseModel):
    """Model capability definition"""
    languages: List[str] = Field(default_factory=list, description="Supported languages")
    frameworks: List[str] = Field(default_factory=list, description="Supported frameworks")
    max_complexity: ComplexityLevel = Field(ComplexityLevel.ADVANCED, description="Max complexity level")
    special_features: List[str] = Field(default_factory=list, description="Special capabilities")
    context_window: int = Field(4096, description="Context window size")
    output_limit: int = Field(4096, description="Max output tokens")


class ModelMetrics(BaseModel):
    """Model performance metrics"""
    average_latency_ms: float = Field(0.0, description="Average latency")
    success_rate: float = Field(1.0, ge=0, le=1, description="Success rate (0-1)")
    cost_per_request: float = Field(0.0, description="Average cost in USD")
    requests_per_minute: int = Field(60, description="Rate limit")
    uptime: float = Field(1.0, ge=0, le=1, description="Uptime percentage")


class ModelConfig(BaseModel):
    """Model configuration"""
    id: str = Field(..., description="Unique model identifier")
    name: str = Field(..., description="Display name")
    provider: ModelProvider = Field(..., description="Provider name")
    model_name: str = Field(..., description="API model name")
    enabled: bool = Field(True, description="Whether model is active")
    priority: int = Field(1, ge=1, le=10, description="Priority for selection")
    capabilities: ModelCapability = Field(default_factory=ModelCapability)
    metrics: ModelMetrics = Field(default_factory=ModelMetrics)
    config: Dict[str, Any] = Field(default_factory=dict, description="Provider-specific config")


class HealthCheck(BaseModel):
    """Health check response"""
    status: str = Field(..., description="Service status")
    version: str = Field(..., description="Service version")
    models_available: int = Field(..., description="Number of available models")
    models_healthy: int = Field(..., description="Number of healthy models")
    uptime_seconds: float = Field(..., description="Service uptime")
    timestamp: datetime = Field(default_factory=datetime.utcnow)


class ErrorResponse(BaseModel):
    """Error response schema"""
    error: str = Field(..., description="Error type")
    message: str = Field(..., description="Error message")
    details: Optional[str] = Field(None, description="Additional details")
    request_id: Optional[str] = Field(None, description="Request ID if available")
    timestamp: datetime = Field(default_factory=datetime.utcnow)
