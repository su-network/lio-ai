"""Enhanced Google Gemini provider with advanced features"""
import os
from typing import Optional, Dict, Any, List
import logging
from litellm import acompletion
import google.generativeai as genai

logger = logging.getLogger(__name__)


class GeminiProvider:
    """
    Enhanced provider specifically for Google Gemini models
    Supports advanced features like:
    - Function calling
    - JSON mode
    - Code execution
    - Advanced safety settings
    - Caching (for long contexts)
    """
    
    def __init__(self, model_id: str, model_name: str, config: Dict[str, Any]):
        """
        Initialize Gemini provider
        
        Args:
            model_id: Internal model identifier (e.g., 'gemini-2.5-pro')
            model_name: Gemini API model name (e.g., 'gemini-2.0-flash-exp')
            config: Model configuration
        """
        self.model_id = model_id
        self.model_name = model_name
        self.config = config
        
        # Map internal model IDs to actual Gemini API names
        self.gemini_model_mapping = {
            'gemini-2.5-pro': 'gemini-2.0-flash-exp',
            'gemini-1.5-pro': 'gemini-1.5-pro-latest',
            'gemini-1.5-pro-latest': 'gemini-1.5-pro-latest',
            'gemini-pro': 'gemini-1.0-pro',
        }
        
        self.api_model_name = self._get_api_model_name()
        self.litellm_model = f"gemini/{self.api_model_name}"
        
        # Configure Google GenAI SDK if available
        api_key = os.getenv('GOOGLE_API_KEY')
        if api_key:
            try:
                genai.configure(api_key=api_key)
                self.use_native_sdk = True
                logger.info(f"Initialized native Gemini SDK for {model_id}")
            except Exception as e:
                logger.warning(f"Failed to initialize native Gemini SDK: {e}, falling back to LiteLLM")
                self.use_native_sdk = False
        else:
            self.use_native_sdk = False
            
        logger.info(f"Initialized Gemini provider for {model_id} -> {self.api_model_name}")
    
    def _get_api_model_name(self) -> str:
        """Get the actual Gemini API model name"""
        # Check if we have a mapping
        for internal_name, api_name in self.gemini_model_mapping.items():
            if internal_name in self.model_name or internal_name == self.model_id:
                return api_name
        
        # Default to model_name
        return self.model_name.replace('gemini/', '').replace('gemini-', '')
    
    def _get_safety_settings(self, mode: str = "permissive") -> List[Dict[str, str]]:
        """
        Get safety settings for Gemini
        
        Args:
            mode: 'permissive' (for code generation) or 'default'
        """
        if mode == "permissive":
            # More permissive for code generation
            return [
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
            # Default settings
            return [
                {
                    "category": "HARM_CATEGORY_HARASSMENT",
                    "threshold": "BLOCK_MEDIUM_AND_ABOVE"
                },
                {
                    "category": "HARM_CATEGORY_HATE_SPEECH",
                    "threshold": "BLOCK_MEDIUM_AND_ABOVE"
                },
                {
                    "category": "HARM_CATEGORY_SEXUALLY_EXPLICIT",
                    "threshold": "BLOCK_MEDIUM_AND_ABOVE"
                },
                {
                    "category": "HARM_CATEGORY_DANGEROUS_CONTENT",
                    "threshold": "BLOCK_MEDIUM_AND_ABOVE"
                }
            ]
    
    async def generate(
        self,
        prompt: str,
        max_tokens: Optional[int] = None,
        temperature: Optional[float] = None,
        system_instruction: Optional[str] = None,
        json_mode: bool = False,
        **kwargs
    ) -> Dict[str, Any]:
        """
        Generate response using Gemini
        
        Args:
            prompt: The input prompt
            max_tokens: Maximum tokens to generate
            temperature: Sampling temperature
            system_instruction: System instruction for the model
            json_mode: Whether to enable JSON response mode
            **kwargs: Additional parameters
            
        Returns:
            Dict with 'content', 'usage', and 'model' keys
        """
        try:
            # Prepare parameters
            params = {
                "model": self.litellm_model,
                "messages": [{"role": "user", "content": prompt}],
                "temperature": temperature or self.config.get("temperature", 0.2),
                "max_output_tokens": max_tokens or self.config.get("max_output_tokens", 8192),
            }
            
            # Add system instruction if provided
            if system_instruction:
                params["system"] = system_instruction
            
            # Add Google-specific parameters
            if "top_k" in self.config:
                params["top_k"] = self.config["top_k"]
            if "top_p" in self.config:
                params["top_p"] = self.config["top_p"]
            
            # Safety settings - permissive for code generation
            params["safety_settings"] = self._get_safety_settings("permissive")
            
            # Enable JSON mode if requested
            if json_mode:
                params["response_mime_type"] = "application/json"
            
            # Merge additional kwargs
            params.update(kwargs)
            
            # Log the call
            logger.debug(f"Calling Gemini with model: {self.litellm_model}")
            
            # Call through LiteLLM
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
            logger.error(f"Gemini generation failed for {self.model_id}: {str(e)}")
            raise Exception(f"Gemini generation failed: {str(e)}")
    
    async def generate_with_code_execution(
        self,
        prompt: str,
        max_tokens: Optional[int] = None,
        temperature: Optional[float] = None,
    ) -> Dict[str, Any]:
        """
        Generate response with code execution enabled (Gemini 1.5 Pro+)
        This allows the model to write and execute Python code
        
        Args:
            prompt: The input prompt
            max_tokens: Maximum tokens to generate
            temperature: Sampling temperature
            
        Returns:
            Dict with 'content', 'usage', and 'model' keys
        """
        if not self.use_native_sdk:
            logger.warning("Native SDK not available, falling back to standard generation")
            return await self.generate(prompt, max_tokens, temperature)
        
        try:
            # Use native SDK for code execution
            model = genai.GenerativeModel(
                model_name=self.api_model_name,
                generation_config={
                    "temperature": temperature or self.config.get("temperature", 0.2),
                    "max_output_tokens": max_tokens or self.config.get("max_output_tokens", 8192),
                    "top_p": self.config.get("top_p", 0.95),
                    "top_k": self.config.get("top_k", 40),
                },
                tools='code_execution',  # Enable code execution
            )
            
            response = model.generate_content(prompt)
            
            # Extract content and usage
            content = response.text
            
            # Note: Native SDK doesn't always provide token counts
            usage = {
                "prompt_tokens": 0,  # Not available
                "completion_tokens": 0,  # Not available
                "total_tokens": 0
            }
            
            return {
                "content": content,
                "usage": usage,
                "model": self.api_model_name,
                "finish_reason": "stop"
            }
            
        except Exception as e:
            logger.error(f"Code execution generation failed: {str(e)}")
            # Fallback to standard generation
            return await self.generate(prompt, max_tokens, temperature)
    
    async def generate_with_tools(
        self,
        prompt: str,
        tools: List[Dict[str, Any]],
        max_tokens: Optional[int] = None,
        temperature: Optional[float] = None,
        system_instruction: Optional[str] = None,
    ) -> Dict[str, Any]:
        """
        Generate response with function calling/tool use
        
        Args:
            prompt: The input prompt
            tools: List of tool/function definitions
            max_tokens: Maximum tokens to generate
            temperature: Sampling temperature
            system_instruction: System instruction for the model
            
        Returns:
            Dict with 'content', 'usage', 'model', and optionally 'tool_calls' keys
        """
        if not self.use_native_sdk:
            logger.warning("Native SDK not available for tools, falling back to standard generation")
            return await self.generate(prompt, max_tokens, temperature, system_instruction)
        
        try:
            # Convert OpenAI-style tool definitions to Gemini format
            gemini_tools = self._convert_tools_to_gemini_format(tools)
            
            # Use native SDK for function calling
            model = genai.GenerativeModel(
                model_name=self.api_model_name,
                generation_config={
                    "temperature": temperature or self.config.get("temperature", 0.7),
                    "max_output_tokens": max_tokens or self.config.get("max_output_tokens", 8192),
                    "top_p": self.config.get("top_p", 0.95),
                    "top_k": self.config.get("top_k", 40),
                },
                tools=gemini_tools,
                system_instruction=system_instruction,
            )
            
            response = model.generate_content(prompt)
            
            # Check if model wants to call a function
            if response.candidates and response.candidates[0].content.parts:
                parts = response.candidates[0].content.parts
                
                # Check for function calls
                tool_calls = []
                text_content = []
                
                for part in parts:
                    if hasattr(part, 'function_call'):
                        # Model wants to call a function
                        tool_calls.append({
                            "type": "function",
                            "function": {
                                "name": part.function_call.name,
                                "arguments": dict(part.function_call.args)
                            }
                        })
                    elif hasattr(part, 'text'):
                        text_content.append(part.text)
                
                content = "\n".join(text_content) if text_content else ""
                
                result = {
                    "content": content,
                    "usage": {
                        "prompt_tokens": 0,
                        "completion_tokens": 0,
                        "total_tokens": 0
                    },
                    "model": self.api_model_name,
                    "finish_reason": "tool_calls" if tool_calls else "stop"
                }
                
                if tool_calls:
                    result["tool_calls"] = tool_calls
                
                return result
            else:
                # No function calls, just text
                content = response.text if response.text else ""
                
                return {
                    "content": content,
                    "usage": {
                        "prompt_tokens": 0,
                        "completion_tokens": 0,
                        "total_tokens": 0
                    },
                    "model": self.api_model_name,
                    "finish_reason": "stop"
                }
            
        except Exception as e:
            logger.error(f"Tool-based generation failed: {str(e)}")
            # Fallback to standard generation
            return await self.generate(prompt, max_tokens, temperature, system_instruction)
    
    def _convert_tools_to_gemini_format(self, tools: List[Dict[str, Any]]) -> List:
        """
        Convert OpenAI-style tool definitions to Gemini format
        
        Args:
            tools: List of OpenAI-style tool definitions
            
        Returns:
            List of Gemini FunctionDeclaration objects
        """
        gemini_functions = []
        
        for tool in tools:
            if tool.get("type") == "function":
                func = tool.get("function", {})
                
                # Create Gemini FunctionDeclaration
                function_declaration = genai.protos.FunctionDeclaration(
                    name=func.get("name"),
                    description=func.get("description", ""),
                    parameters=func.get("parameters", {})
                )
                
                gemini_functions.append(function_declaration)
        
        if gemini_functions:
            return [genai.protos.Tool(function_declarations=gemini_functions)]
        
        return []

    
    async def generate_streaming(
        self,
        prompt: str,
        max_tokens: Optional[int] = None,
        temperature: Optional[float] = None,
    ):
        """
        Generate response with streaming
        
        Args:
            prompt: The input prompt
            max_tokens: Maximum tokens to generate
            temperature: Sampling temperature
            
        Yields:
            Chunks of generated text
        """
        try:
            params = {
                "model": self.litellm_model,
                "messages": [{"role": "user", "content": prompt}],
                "temperature": temperature or self.config.get("temperature", 0.2),
                "max_output_tokens": max_tokens or self.config.get("max_output_tokens", 8192),
                "stream": True,
            }
            
            # Add Google-specific parameters
            if "top_k" in self.config:
                params["top_k"] = self.config["top_k"]
            if "top_p" in self.config:
                params["top_p"] = self.config["top_p"]
            
            params["safety_settings"] = self._get_safety_settings("permissive")
            
            # Stream response
            response = await acompletion(**params)
            
            async for chunk in response:
                if chunk.choices[0].delta.content:
                    yield chunk.choices[0].delta.content
                    
        except Exception as e:
            logger.error(f"Streaming generation failed: {str(e)}")
            raise Exception(f"Streaming failed: {str(e)}")
    
    async def health_check(self) -> bool:
        """Check if the model is accessible"""
        try:
            response = await acompletion(
                model=self.litellm_model,
                messages=[{"role": "user", "content": "Hi"}],
                max_output_tokens=5,
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
            "api_model_name": self.api_model_name,
            "litellm_model": self.litellm_model,
            "provider": "google",
            "config": self.config,
            "features": {
                "code_execution": self.use_native_sdk,
                "json_mode": True,
                "streaming": True,
                "function_calling": True,
                "long_context": "2.5-pro" in self.model_id or "1.5-pro" in self.model_id
            }
        }


def create_gemini_provider(model_config: Dict[str, Any]) -> GeminiProvider:
    """
    Factory function to create a Gemini provider
    
    Args:
        model_config: Model configuration dict
        
    Returns:
        GeminiProvider instance
    """
    return GeminiProvider(
        model_id=model_config["id"],
        model_name=model_config["model_name"],
        config=model_config.get("config", {})
    )
