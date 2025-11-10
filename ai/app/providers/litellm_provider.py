"""LiteLLM-based unified provider for all LLM models"""
import os
from typing import Optional, Dict, Any, List
import litellm
from litellm import completion, acompletion
import logging

logger = logging.getLogger(__name__)

# Configure LiteLLM
litellm.drop_params = True  # Drop unsupported params instead of erroring
litellm.set_verbose = False


class LiteLLMProvider:
    """Unified provider using LiteLLM for all model integrations"""
    
    def __init__(self, model_id: str, model_name: str, provider: str, config: Dict[str, Any]):
        """
        Initialize LiteLLM provider
        
        Args:
            model_id: Internal model identifier (e.g., 'gpt-4-turbo')
            model_name: LiteLLM model name (e.g., 'gpt-4-turbo-preview')
            provider: Provider name (openai, anthropic, google, cohere, etc.)
            config: Model configuration (temperature, max_tokens, etc.)
        """
        self.model_id = model_id
        self.model_name = model_name
        self.provider = provider
        self.config = config
        
        # Map provider to LiteLLM model format
        self.litellm_model = self._get_litellm_model_name()
        
        logger.info(f"Initialized LiteLLM provider for {model_id} ({self.litellm_model})")
    
    def _get_litellm_model_name(self) -> str:
        """Convert model name to LiteLLM format"""
        provider_prefixes = {
            'anthropic': 'claude',
            'google': 'gemini',
            'cohere': 'command',
            'openai': ''  # OpenAI models don't need prefix
        }
        
        # Special handling for Google/Gemini models
        if self.provider == 'google':
            # Map our internal names to actual Gemini API model names
            gemini_model_mapping = {
                'gemini-2.5-pro': 'gemini-2.0-flash-exp',  # Using latest available
                'gemini-1.5-pro': 'gemini-1.5-pro-latest',
                'gemini-pro': 'gemini-1.0-pro'
            }
            
            # Check if we have a mapping for this model
            for internal_name, api_name in gemini_model_mapping.items():
                if internal_name in self.model_name or internal_name == self.model_id:
                    return f"gemini/{api_name}"
            
            # Default: use the model_name as-is with gemini prefix
            if not self.model_name.startswith('gemini/'):
                return f"gemini/{self.model_name}"
            return self.model_name
        
        # If model already starts with provider prefix, return as-is
        if self.provider in provider_prefixes and self.model_name.startswith(provider_prefixes[self.provider]):
            return self.model_name
        
        # For OpenAI, use model name directly
        if self.provider == 'openai':
            return self.model_name
        
        return self.model_name
    
    async def generate(
        self,
        prompt: str,
        max_tokens: Optional[int] = None,
        temperature: Optional[float] = None,
        **kwargs
    ) -> Dict[str, Any]:
        """
        Generate code using LiteLLM
        
        Args:
            prompt: The input prompt
            max_tokens: Maximum tokens to generate
            temperature: Sampling temperature
            **kwargs: Additional parameters
            
        Returns:
            Dict with 'content', 'usage', and 'model' keys
        """
        try:
            # Prepare parameters
            params = {
                "model": self.litellm_model,
                "messages": [{"role": "user", "content": prompt}],
                "temperature": temperature or self.config.get("temperature", 0.7),
            }
            
            # Handle max_tokens/max_output_tokens for different providers
            max_token_value = max_tokens or self.config.get("max_output_tokens") or self.config.get("max_tokens", 2048)
            
            if self.provider == "google":
                # Google uses max_output_tokens instead of max_tokens
                params["max_output_tokens"] = max_token_value
                
                # Add Google-specific parameters
                if "top_k" in self.config:
                    params["top_k"] = self.config["top_k"]
                if "top_p" in self.config:
                    params["top_p"] = self.config["top_p"]
                    
                # Set safety settings for Gemini to be less restrictive for code generation
                params["safety_settings"] = [
                    {
                        "category": "HARM_CATEGORY_HARASSMENT",
                        "threshold": "BLOCK_NONE"
                    },
                    {
                        "category": "HARM_CATEGORY_HATE_SPEECH",
                        "threshold": "BLOCK_NONE"
                    },
                    {
                        "category": "HARM_CATEGORY_SEXUALLY_EXPLICIT",
                        "threshold": "BLOCK_NONE"
                    },
                    {
                        "category": "HARM_CATEGORY_DANGEROUS_CONTENT",
                        "threshold": "BLOCK_NONE"
                    }
                ]
            else:
                params["max_tokens"] = max_token_value
                
                # Add provider-specific params
                if "top_p" in self.config:
                    params["top_p"] = self.config["top_p"]
            
            # Merge additional kwargs
            params.update(kwargs)
            
            # Call LiteLLM
            logger.debug(f"Calling LiteLLM with model: {self.litellm_model}, params: {params}")
            response = await acompletion(**params)
            
            # Extract response
            content = response.choices[0].message.content
            usage = {
                "prompt_tokens": response.usage.prompt_tokens,
                "completion_tokens": response.usage.completion_tokens,
                "total_tokens": response.usage.total_tokens
            }
            
            return {
                "content": content,
                "usage": usage,
                "model": self.litellm_model,
                "finish_reason": response.choices[0].finish_reason
            }
            
        except Exception as e:
            logger.error(f"LiteLLM generation failed for {self.model_id}: {str(e)}")
            raise Exception(f"Generation failed: {str(e)}")
    
    async def health_check(self) -> bool:
        """
        Check if the model is accessible and healthy
        
        Returns:
            True if healthy, False otherwise
        """
        try:
            # Try a minimal completion
            response = await acompletion(
                model=self.litellm_model,
                messages=[{"role": "user", "content": "Hi"}],
                max_tokens=5,
                timeout=10
            )
            return response is not None
        except Exception as e:
            logger.debug(f"Health check failed for {self.model_id}: {str(e)}")
            return False
    
    def get_info(self) -> Dict[str, Any]:
        """Get provider information"""
        return {
            "model_id": self.model_id,
            "model_name": self.model_name,
            "litellm_model": self.litellm_model,
            "provider": self.provider,
            "config": self.config
        }


def create_provider(model_config: Dict[str, Any]) -> LiteLLMProvider:
    """
    Factory function to create a LiteLLM provider
    
    Args:
        model_config: Model configuration dict
        
    Returns:
        LiteLLMProvider instance
    """
    return LiteLLMProvider(
        model_id=model_config["id"],
        model_name=model_config["model_name"],
        provider=model_config["provider"],
        config=model_config.get("config", {})
    )
