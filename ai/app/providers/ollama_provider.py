"""Ollama provider for local model inference"""
import os
import logging
from typing import Optional, Dict, Any, AsyncGenerator, List
import httpx

logger = logging.getLogger(__name__)


class OllamaProvider:
    """
    Provider for Ollama - run local LLM models
    No API key required, connects to local Ollama instance
    """
    
    def __init__(self, model_id: str, model_name: str, config: Dict[str, Any]):
        """
        Initialize Ollama provider
        
        Args:
            model_id: Internal model identifier (e.g., 'llama3')
            model_name: Ollama model name (e.g., 'llama3:latest')
            config: Model configuration
        """
        self.model_id = model_id
        self.model_name = model_name
        self.config = config
        
        # Get Ollama base URL from environment or use default
        self.base_url = os.getenv('OLLAMA_BASE_URL', 'http://localhost:11434')
        
        logger.info(f"Initialized Ollama provider for {model_id} -> {model_name} at {self.base_url}")
    
    async def generate(
        self,
        prompt: str,
        max_tokens: Optional[int] = None,
        temperature: Optional[float] = None,
        system_instruction: Optional[str] = None,
        **kwargs
    ) -> Dict[str, Any]:
        """
        Generate response using Ollama
        
        Args:
            prompt: The input prompt (will be converted to messages format)
            max_tokens: Maximum tokens to generate
            temperature: Sampling temperature
            system_instruction: System instruction for the model
            **kwargs: Additional parameters
            
        Returns:
            Dict with 'content', 'usage', and 'model' keys
        """
        try:
            # Convert prompt to messages format
            messages = []
            if system_instruction:
                messages.append({"role": "system", "content": system_instruction})
            messages.append({"role": "user", "content": prompt})
            
            # Prepare request
            payload = {
                "model": self.model_name,
                "messages": messages,
                "stream": False,
                "options": {}
            }
            
            # Add parameters
            if temperature is not None:
                payload["options"]["temperature"] = temperature
            elif "temperature" in self.config:
                payload["options"]["temperature"] = self.config["temperature"]
            
            if max_tokens:
                payload["options"]["num_predict"] = max_tokens
            
            # Add any additional options
            if "top_p" in self.config:
                payload["options"]["top_p"] = self.config["top_p"]
            if "top_k" in self.config:
                payload["options"]["top_k"] = self.config["top_k"]
            
            # Make request to Ollama
            async with httpx.AsyncClient(timeout=120.0) as client:
                response = await client.post(
                    f"{self.base_url}/api/chat",
                    json=payload
                )
                response.raise_for_status()
                result = response.json()
            
            # Extract response
            content = result.get("message", {}).get("content", "")
            
            # Build usage info (Ollama provides this)
            usage = {
                "prompt_tokens": result.get("prompt_eval_count", 0),
                "completion_tokens": result.get("eval_count", 0),
                "total_tokens": result.get("prompt_eval_count", 0) + result.get("eval_count", 0)
            }
            
            return {
                "content": content,
                "usage": usage,
                "model": self.model_name,
                "finish_reason": "stop"
            }
            
        except httpx.ConnectError as e:
            logger.error(f"Failed to connect to Ollama at {self.base_url}: {str(e)}")
            raise Exception(f"Ollama is not running. Please start Ollama on your device. Visit https://ollama.ai to download.")
        except httpx.HTTPStatusError as e:
            logger.error(f"Ollama HTTP error: {e.response.status_code} - {e.response.text}")
            if e.response.status_code == 404:
                raise Exception(f"Model '{self.model_name}' not found. Run 'ollama pull {self.model_name}' to download it.")
            raise Exception(f"Ollama error: {str(e)}")
        except Exception as e:
            logger.error(f"Ollama generation failed for {self.model_id}: {str(e)}")
            raise Exception(f"Ollama generation failed: {str(e)}")
    
    async def generate_streaming(
        self,
        messages: list,
        max_tokens: Optional[int] = None,
        temperature: Optional[float] = None,
        **kwargs
    ) -> AsyncGenerator[str, None]:
        """
        Generate streaming response using Ollama
        
        Args:
            messages: List of message dicts
            max_tokens: Maximum tokens to generate
            temperature: Sampling temperature
            **kwargs: Additional parameters
            
        Yields:
            Response chunks as they arrive
        """
        try:
            payload = {
                "model": self.model_name,
                "messages": messages,
                "stream": True,
                "options": {}
            }
            
            if temperature is not None:
                payload["options"]["temperature"] = temperature
            elif "temperature" in self.config:
                payload["options"]["temperature"] = self.config["temperature"]
            
            if max_tokens:
                payload["options"]["num_predict"] = max_tokens
            
            async with httpx.AsyncClient(timeout=120.0) as client:
                async with client.stream(
                    "POST",
                    f"{self.base_url}/api/chat",
                    json=payload
                ) as response:
                    response.raise_for_status()
                    async for line in response.aiter_lines():
                        if line:
                            import json
                            chunk = json.loads(line)
                            if "message" in chunk and "content" in chunk["message"]:
                                yield chunk["message"]["content"]
                            
                            # Check if done
                            if chunk.get("done", False):
                                break
        
        except Exception as e:
            logger.error(f"Ollama streaming failed: {str(e)}")
            raise
    
    async def health_check(self) -> bool:
        """Check if Ollama is running and model is available"""
        try:
            async with httpx.AsyncClient(timeout=10.0) as client:
                # Check if Ollama is running
                response = await client.get(f"{self.base_url}/api/tags")
                response.raise_for_status()
                
                # Check if our specific model is available
                models = response.json().get("models", [])
                model_names = [m.get("name", "") for m in models]
                
                # Check if model exists (with or without tag)
                base_model = self.model_name.split(":")[0]
                return any(base_model in name for name in model_names)
                
        except Exception as e:
            logger.debug(f"Ollama health check failed: {str(e)}")
            return False
    
    @staticmethod
    async def list_available_models(base_url: str = None) -> List[Dict[str, Any]]:
        """
        List all available models from Ollama
        
        Args:
            base_url: Ollama base URL (defaults to localhost:11434)
            
        Returns:
            List of model dictionaries with name, size, and parameter info
        """
        if base_url is None:
            base_url = os.getenv('OLLAMA_BASE_URL', 'http://localhost:11434')
        
        try:
            async with httpx.AsyncClient(timeout=10.0) as client:
                response = await client.get(f"{base_url}/api/tags")
                response.raise_for_status()
                
                models_data = response.json().get("models", [])
                
                # Format model information
                available_models = []
                for model in models_data:
                    model_name = model.get("name", "")
                    details = model.get("details", {})
                    
                    # Extract parameter size (e.g., "8.0B")
                    param_size = details.get("parameter_size", "")
                    
                    # Create friendly display name
                    base_name = model_name.split(":")[0]
                    display_name = base_name.replace("-", " ").title()
                    if param_size:
                        display_name = f"{display_name} ({param_size})"
                    
                    available_models.append({
                        "id": base_name,  # e.g., "llama3"
                        "name": display_name,  # e.g., "Llama3 (8.0B)"
                        "model_name": model_name,  # e.g., "llama3:latest"
                        "size": model.get("size", 0),
                        "parameter_size": param_size,
                        "family": details.get("family", ""),
                        "quantization": details.get("quantization_level", "")
                    })
                
                logger.info(f"Found {len(available_models)} Ollama models: {[m['id'] for m in available_models]}")
                return available_models
                
        except Exception as e:
            logger.warning(f"Failed to list Ollama models: {str(e)}")
            return []
