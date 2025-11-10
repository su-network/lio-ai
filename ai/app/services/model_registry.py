"""Model registry service for managing available LLM models"""
from typing import List, Dict, Optional, Union
from app.models import ModelConfig, ModelProvider, ComplexityLevel
from app.providers.litellm_provider import LiteLLMProvider
from app.providers.gemini_provider import GeminiProvider, create_gemini_provider
from app.config import settings
import yaml
import logging
import os
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
        self._load_config()
    
    def _check_api_key_available(self, provider: str) -> bool:
        """Check if API key is configured for a provider"""
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
    
    def _load_config(self):
        """Load models configuration from YAML"""
        try:
            with open(self.config_path, 'r') as f:
                config = yaml.safe_load(f)
            
            # Load models
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
            
            # Load selection strategies
            self.selection_strategies = config.get('selection_strategies', {})
            
            # Load fallback configuration
            self.fallback_config = config.get('fallback', {})
            
            logger.info(f"Loaded {len(self.models)} models, {len(self.providers)} providers active")
            
        except Exception as e:
            logger.error(f"Failed to load model config: {str(e)}")
            raise
    
    def _create_provider(self, model_config: ModelConfig) -> Optional[Union[LiteLLMProvider, GeminiProvider]]:
        """Create provider instance - uses enhanced Gemini provider for Google models"""
        try:
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
    
    def reload_config(self):
        """Reload configuration from file"""
        self.models.clear()
        self.providers.clear()
        self._load_config()
        logger.info("Configuration reloaded")
