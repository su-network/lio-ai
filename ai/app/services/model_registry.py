"""Model registry service for managing available LLM models"""
from typing import List, Dict, Optional, Union
from app.models import ModelConfig, ModelProvider, ComplexityLevel
from app.providers.litellm_provider import LiteLLMProvider
from app.providers.gemini_provider import GeminiProvider, create_gemini_provider
from app.providers.ollama_provider import OllamaProvider
from app.config import settings
import yaml
import logging
import os
import asyncio
from pathlib import Path

logger = logging.getLogger(__name__)


class ModelRegistry:
    """Registry for managing and selecting LLM models"""
    
    def __init__(self, config_path: str):
        self.config_path = config_path
        self.models: Dict[str, ModelConfig] = {}
        self.providers: Dict[str, Union[LiteLLMProvider, GeminiProvider]] = {}
        self.selection_strategies: Dict = {}
        self.fallback_config: Dict = {}
        self.model_health_status: Dict[str, bool] = {}
        # Don't call _load_config here, call initialize() from async context
    
    async def initialize(self):
        """Initialize the registry - must be called in async context"""
        await self._load_config()
    
    def _check_api_key_available(self, provider: str) -> bool:
        """Check if API key is configured for a provider"""
        # Ollama doesn't need an API key - it's local
        if provider.lower() == 'ollama':
            return True
        
        key_mapping = {
            'openai': settings.openai_api_key,
            'anthropic': settings.anthropic_api_key,
            'google': settings.google_api_key,
            'cohere': settings.cohere_api_key,
        }
        
        api_key = key_mapping.get(provider.lower())
        has_key = bool(api_key and len(api_key) > 0)
        
        if not has_key:
            logger.warning(f"No API key configured for provider: {provider}")
        
        return has_key
    
    async def _load_config(self):
        """Load models configuration from YAML"""
        try:
            with open(self.config_path, 'r') as f:
                config = yaml.safe_load(f)
            
            # Load regular models
            for model_data in config.get('models', []):
                model_config = ModelConfig(**model_data)
                self.models[model_config.id] = model_config
                
                # Check if API key is available
                has_api_key = self._check_api_key_available(model_config.provider.value)
                
                # Only initialize provider if enabled AND has API key
                if model_config.enabled and has_api_key:
                    provider = self._create_provider(model_config)
                    if provider:
                        self.providers[model_config.id] = provider
                        logger.info(f"✓ Initialized provider for {model_config.id}")
                    else:
                        logger.warning(f"✗ Failed to initialize provider for {model_config.id}")
                        self.model_health_status[model_config.id] = False
                elif not has_api_key:
                    logger.info(f"⊘ Skipping {model_config.id} - no API key for {model_config.provider.value}")
                    self.model_health_status[model_config.id] = False
            
            # Load Ollama models dynamically from local Ollama instance
            await self._load_ollama_models()
            
            # Load selection strategies
            self.selection_strategies = config.get('selection_strategies', {})
            
            # Load fallback configuration
            self.fallback_config = config.get('fallback', {})
            
            logger.info(f"Loaded {len(self.models)} models, {len(self.providers)} providers active")
            
        except Exception as e:
            logger.error(f"Failed to load model config: {str(e)}")
            raise
    
    async def _load_ollama_models(self):
        """Dynamically load models from local Ollama instance"""
        try:
            logger.info("Detecting Ollama models...")
            available_models = await OllamaProvider.list_available_models()
            
            if not available_models:
                logger.warning("No Ollama models detected. Install models with: ollama pull <model-name>")
                return
            
            # Model-specific capabilities based on specialization
            model_capabilities = {
                "codegemma": {
                    "languages": ["python", "javascript", "typescript", "java", "go", "rust", "cpp", "kotlin", "swift"],
                    "frameworks": ["all"],
                    "max_complexity": "advanced",
                    "special_features": ["code-generation", "code-completion", "code-analysis", "refactoring", "debugging"],
                    "context_window": 8192,
                    "output_limit": 4096,
                    "description": "🏆 Best for code: Google's code specialist with strong code understanding"
                },
                "qwen2.5-coder": {
                    "languages": ["python", "javascript", "typescript", "java", "go", "rust", "cpp", "c", "php", "ruby"],
                    "frameworks": ["all"],
                    "max_complexity": "advanced",
                    "special_features": ["code-generation", "code-completion", "code-analysis", "bug-fixing", "optimization"],
                    "context_window": 32768,  # Qwen has large context
                    "output_limit": 8192,
                    "description": "🚀 Excellent for code: Alibaba's specialized coder with 32K context window"
                },
                "codellama": {
                    "languages": ["python", "javascript", "typescript", "java", "cpp", "go", "rust", "bash"],
                    "frameworks": ["all"],
                    "max_complexity": "advanced",
                    "special_features": ["code-generation", "code-infilling", "code-completion", "instruction-following"],
                    "context_window": 16384,
                    "output_limit": 4096,
                    "description": "🔥 Meta's code specialist: Excellent for complex programming tasks"
                },
                "deepseek-coder": {
                    "languages": ["python", "javascript", "typescript", "java", "cpp", "go", "rust", "c"],
                    "frameworks": ["all"],
                    "max_complexity": "advanced",
                    "special_features": ["code-generation", "code-completion", "algorithm-design", "system-design"],
                    "context_window": 16384,
                    "output_limit": 4096,
                    "description": "💎 Strong coder: DeepSeek's specialized model for programming"
                },
                "phi3": {
                    "languages": ["python", "javascript", "typescript", "java", "go"],
                    "frameworks": ["fastapi", "express", "gin", "django", "flask"],
                    "max_complexity": "intermediate",
                    "special_features": ["chat", "code-generation", "quick-answers"],
                    "context_window": 4096,
                    "output_limit": 2048,
                    "description": "⚡ Fast & lightweight: Good for simple tasks and quick responses"
                }
            }
            
            # Default capabilities for unknown models
            default_capabilities = {
                "languages": ["python", "javascript", "typescript", "java", "go", "rust", "cpp"],
                "frameworks": ["all"],
                "max_complexity": "advanced",
                "special_features": ["code-generation", "code-analysis", "chat"],
                "context_window": 4096,
                "output_limit": 4096,
                "description": "General purpose coding model"
            }
            
            # Create model configs for each detected model
            for model_info in available_models:
                model_id = model_info["id"]
                
                # Get model-specific capabilities or use defaults
                capabilities_data = model_capabilities.get(model_id, default_capabilities).copy()
                description = capabilities_data.pop("description", "General purpose model")
                
                # Set priority based on model type (code specialists get higher priority)
                priority = 1 if "code" in model_id or model_id in ["codegemma", "qwen2.5-coder"] else 3
                
                # Create model config
                model_config = ModelConfig(
                    id=model_id,
                    name=model_info["name"],
                    provider=ModelProvider.OLLAMA,
                    model_name=model_info["model_name"],
                    enabled=True,
                    priority=priority,
                    capabilities=capabilities_data,
                    metrics={
                        "average_latency_ms": 0.0,
                        "success_rate": 1.0,
                        "cost_per_request": 0.0,  # Free!
                        "requests_per_minute": 60,
                        "uptime": 1.0
                    },
                    config={"description": description}  # Store description in config
                )
                
                # Register the model
                self.models[model_id] = model_config
                
                # Create provider
                provider = self._create_provider(model_config)
                if provider:
                    self.providers[model_id] = provider
                    logger.info(f"✓ Registered Ollama model: {model_info['name']} ({model_info['model_name']})")
                else:
                    logger.warning(f"✗ Failed to create provider for Ollama model: {model_id}")
            
            logger.info(f"Loaded {len(available_models)} Ollama models")
            
        except Exception as e:
            logger.warning(f"Failed to load Ollama models: {str(e)}")
    
    def _create_provider(self, model_config: ModelConfig) -> Optional[Union[LiteLLMProvider, GeminiProvider, OllamaProvider]]:
        """Create provider instance - uses enhanced Gemini provider for Google models, OllamaProvider for local models"""
        try:
            # Use Ollama provider for local models
            if model_config.provider.value.lower() == 'ollama':
                logger.info(f"Creating Ollama provider for {model_config.id}")
                return OllamaProvider(
                    model_id=model_config.id,
                    model_name=model_config.model_name,
                    config=model_config.config
                )
            
            # Use enhanced Gemini provider for Google models
            if model_config.provider == ModelProvider.GOOGLE:
                logger.info(f"Creating enhanced Gemini provider for {model_config.id}")
                return create_gemini_provider({
                    "id": model_config.id,
                    "model_name": model_config.model_name,
                    "config": model_config.config
                })
            
            # Use LiteLLM for other providers
            return LiteLLMProvider(
                model_id=model_config.id,
                model_name=model_config.model_name,
                provider=model_config.provider.value,
                config=model_config.config
            )
        except Exception as e:
            logger.error(f"Failed to create provider for {model_config.id}: {str(e)}")
            return None
    
    def get_model(self, model_id: str) -> Optional[ModelConfig]:
        """Get model configuration by ID"""
        return self.models.get(model_id)
    
    def get_provider(self, model_id: str) -> Optional[Union[LiteLLMProvider, GeminiProvider]]:
        """Get provider instance by model ID"""
        return self.providers.get(model_id)
    
    def list_models(
        self, 
        language: Optional[str] = None,
        complexity: Optional[ComplexityLevel] = None,
        enabled_only: bool = True
    ) -> List[ModelConfig]:
        """List models with optional filtering"""
        models = list(self.models.values())
        
        if enabled_only:
            # Only return models that are both enabled in config AND have providers initialized
            models = [m for m in models if m.enabled and m.id in self.providers]
        
        if language:
            models = [m for m in models if language in m.capabilities.languages]
        
        if complexity:
            complexity_order = {
                ComplexityLevel.SIMPLE: 0,
                ComplexityLevel.INTERMEDIATE: 1,
                ComplexityLevel.ADVANCED: 2
            }
            required_level = complexity_order[complexity]
            models = [
                m for m in models 
                if complexity_order[m.capabilities.max_complexity] >= required_level
            ]
        
        return models
    
    def select_models(
        self,
        language: str,
        complexity: ComplexityLevel,
        framework: Optional[str] = None,
        max_models: int = 3,
        strategy: str = "default"
    ) -> List[str]:
        """Intelligently select best models for the request"""
        
        # Filter eligible models
        eligible = self.list_models(language=language, complexity=complexity)
        
        if not eligible:
            return []
        
        # Get selection strategy
        strategy_config = self.selection_strategies.get(
            strategy, 
            self.selection_strategies.get("default", {})
        )
        factors = strategy_config.get("factors", {})
        
        # Score each model
        scored_models = []
        for model in eligible:
            score = self._calculate_model_score(
                model, language, framework, complexity, factors
            )
            scored_models.append((model.id, score))
        
        # Sort by score descending
        scored_models.sort(key=lambda x: x[1], reverse=True)
        
        # Return top N model IDs
        return [model_id for model_id, _ in scored_models[:max_models]]
    
    def _calculate_model_score(
        self,
        model: ModelConfig,
        language: str,
        framework: Optional[str],
        complexity: ComplexityLevel,
        factors: Dict[str, float]
    ) -> float:
        """Calculate model suitability score"""
        score = 0.0
        
        # Language match
        if language in model.capabilities.languages:
            score += factors.get("language_match", 0.3)
        
        # Framework match
        if framework and framework in model.capabilities.frameworks:
            score += factors.get("framework_match", 0.2)
        
        # Complexity match
        complexity_order = {
            ComplexityLevel.SIMPLE: 0,
            ComplexityLevel.INTERMEDIATE: 1,
            ComplexityLevel.ADVANCED: 2
        }
        if complexity_order[model.capabilities.max_complexity] >= complexity_order[complexity]:
            score += factors.get("complexity_match", 0.2)
        
        # Cost efficiency (inverse of cost)
        cost_score = 1.0 - min(model.metrics.cost_per_request / 0.10, 1.0)
        score += cost_score * factors.get("cost_efficiency", 0.15)
        
        # Speed (inverse of latency)
        speed_score = 1.0 - min(model.metrics.average_latency_ms / 5000, 1.0)
        score += speed_score * factors.get("speed", 0.10)
        
        # Success rate
        score += model.metrics.success_rate * factors.get("success_rate", 0.05)
        
        # Priority boost
        score += (model.priority / 10.0) * 0.1
        
        return score
    
    def get_fallback_models(self) -> List[str]:
        """Get fallback model order"""
        return self.fallback_config.get("fallback_order", [])
    
    def get_model_status(self, model_id: str) -> Dict[str, any]:
        """Get detailed status of a specific model"""
        model = self.models.get(model_id)
        if not model:
            return {"status": "not_found"}
        
        has_api_key = self._check_api_key_available(model.provider.value)
        has_provider = model_id in self.providers
        is_healthy = self.model_health_status.get(model_id, False)
        
        status = "unavailable"
        reason = None
        
        if not model.enabled:
            status = "disabled"
            reason = "Model disabled in configuration"
        elif not has_api_key:
            status = "no_api_key"
            reason = f"Missing API key for {model.provider.value}"
        elif not has_provider:
            status = "not_initialized"
            reason = "Provider initialization failed"
        elif has_provider:
            status = "available"
            reason = "Model ready to use"
        
        return {
            "model_id": model_id,
            "status": status,
            "reason": reason,
            "has_api_key": has_api_key,
            "has_provider": has_provider,
            "is_healthy": is_healthy,
            "provider": model.provider.value
        }
    
    def get_all_models_status(self) -> List[Dict[str, any]]:
        """Get status of all models"""
        return [self.get_model_status(model_id) for model_id in self.models.keys()]
    
    async def health_check_all(self) -> Dict[str, bool]:
        """Check health of all active providers"""
        health_status = {}
        for model_id, provider in self.providers.items():
            try:
                is_healthy = await provider.health_check()
                health_status[model_id] = is_healthy
                self.model_health_status[model_id] = is_healthy
            except Exception as e:
                logger.error(f"Health check failed for {model_id}: {str(e)}")
                health_status[model_id] = False
                self.model_health_status[model_id] = False
        return health_status
    
    async def reload_config(self):
        """Reload configuration from file"""
        self.models.clear()
        self.providers.clear()
        await self._load_config()
        logger.info("Configuration reloaded")
