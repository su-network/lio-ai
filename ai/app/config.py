"""Configuration management for Lio AI service"""
from pydantic_settings import BaseSettings
from typing import Optional
from pathlib import Path


class Settings(BaseSettings):
    """Application settings - reads from root .env file"""
    
    # Application (prefixed with AI_ in .env)
    ai_app_name: str = "Lio AI Service"
    ai_app_version: str = "0.1.0"
    ai_debug: bool = False
    
    # Server
    ai_host: str = "0.0.0.0"
    ai_port: int = 8000
    
    # API Keys
    openai_api_key: Optional[str] = None
    anthropic_api_key: Optional[str] = None
    google_api_key: Optional[str] = None
    cohere_api_key: Optional[str] = None
    
    # Redis
    redis_url: str = "redis://localhost:6379"
    cache_ttl: int = 3600  # 1 hour
    
    # Model Configuration
    default_temperature: float = 0.2
    default_max_tokens: int = 4096
    default_timeout: int = 30
    
    # Rate Limiting
    max_requests_per_minute: int = 60
    max_concurrent_requests: int = 10
    
    # Prompts Configuration
    prompts_config_path: str = "config/prompts.yaml"
    models_config_path: str = "config/models.yaml"
    
    # Computed properties
    @property
    def app_name(self) -> str:
        return self.ai_app_name
    
    @property
    def app_version(self) -> str:
        return self.ai_app_version
    
    @property
    def debug(self) -> bool:
        return self.ai_debug
    
    @property
    def host(self) -> str:
        return self.ai_host
    
    @property
    def port(self) -> int:
        return self.ai_port
    
    class Config:
        # Read from root .env file
        env_file = "../.env"
        env_file_encoding = "utf-8"
        case_sensitive = False
        extra = "ignore"  # Ignore extra fields from shared .env


settings = Settings()
