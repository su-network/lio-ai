"""Code generation orchestration service"""
import asyncio
from typing import List, Dict, Optional
from app.models import CodeGenRequest, CodeGenResponse, ModelResponse, QualityMetrics
from app.services.model_registry import ModelRegistry
from app.services.prompt_manager import PromptManager
from tenacity import retry, stop_after_attempt, wait_exponential
import time
import logging

logger = logging.getLogger(__name__)


class CodeGenerationService:
    """Main service for orchestrating multi-model code generation"""
    
    def __init__(self, model_registry: ModelRegistry, prompt_manager: PromptManager):
        self.model_registry = model_registry
        self.prompt_manager = prompt_manager
    
    async def generate_code(self, request: CodeGenRequest) -> CodeGenResponse:
        """
        Generate code using multiple models in parallel
        """
        start_time = time.time()
        
        try:
            # Auto-select models if not specified or validate selected models
            model_ids = await self._prepare_model_selection(request)
            
            if not model_ids:
                return self._create_error_response(
                    request,
                    "No suitable models available for this request",
                    time.time() - start_time
                )
            
            # Build prompt
            prompt = self.prompt_manager.build_prompt(
                language=request.language,
                complexity=request.complexity.value,
                prompt=request.prompt,
                framework=request.framework,
                context=request.context,
                constraints=request.constraints,
                include_tests=request.include_tests,
                include_docs=request.include_docs
            )
            
            logger.info(f"Generating code with models: {model_ids}")
            
            # Execute models in parallel
            model_responses = await self._execute_parallel(request, model_ids, prompt)
            
            # Calculate total time
            total_time = time.time() - start_time
            
            # Determine status
            successful_responses = [r for r in model_responses if not r.error]
            if not successful_responses:
                status = "failed"
            elif len(successful_responses) < len(model_responses):
                status = "partial"
            else:
                status = "success"
            
            # Generate consensus code if multiple successful responses
            consensus_code = None
            best_model = None
            if len(successful_responses) > 1:
                consensus_code, best_model = self._generate_consensus(successful_responses)
            elif len(successful_responses) == 1:
                best_model = successful_responses[0].model_id
            
            return CodeGenResponse(
                request_id=request.request_id,
                status=status,
                total_time_ms=total_time * 1000,
                model_responses=model_responses,
                consensus_code=consensus_code,
                best_model=best_model,
                metadata={
                    "language": request.language,
                    "complexity": request.complexity.value,
                    "models_attempted": len(model_ids),
                    "models_succeeded": len(successful_responses)
                }
            )
            
        except Exception as e:
            logger.error(f"Code generation failed: {str(e)}")
            return self._create_error_response(
                request,
                str(e),
                time.time() - start_time
            )
    
    async def _prepare_model_selection(self, request: CodeGenRequest) -> List[str]:
        """Prepare list of models to use"""
        
        if request.selected_models and request.selected_models[0] != "auto":
            # Validate selected models
            validated_models = []
            for model_id in request.selected_models:
                provider = self.model_registry.get_provider(model_id)
                if provider and provider.enabled:
                    validated_models.append(model_id)
                else:
                    logger.warning(f"Model {model_id} not available, skipping")
            return validated_models
        else:
            # Auto-select models
            max_models = request.max_models or 3
            return self.model_registry.select_models(
                language=request.language,
                complexity=request.complexity,
                framework=request.framework,
                max_models=max_models
            )
    
    async def _execute_parallel(
        self, 
        request: CodeGenRequest, 
        model_ids: List[str], 
        prompt: str
    ) -> List[ModelResponse]:
        """Execute multiple models in parallel with timeout"""
        
        tasks = []
        for model_id in model_ids:
            provider = self.model_registry.get_provider(model_id)
            if provider:
                task = self._execute_with_fallback(request, model_id, provider, prompt)
                tasks.append(task)
        
        # Execute all tasks with timeout
        try:
            responses = await asyncio.wait_for(
                asyncio.gather(*tasks, return_exceptions=True),
                timeout=request.timeout
            )
            
            # Filter out exceptions
            valid_responses = []
            for response in responses:
                if isinstance(response, ModelResponse):
                    valid_responses.append(response)
                elif isinstance(response, Exception):
                    logger.error(f"Task failed with exception: {str(response)}")
            
            return valid_responses
            
        except asyncio.TimeoutError:
            logger.error("Code generation timeout")
            return []
    
    async def _execute_with_fallback(
        self,
        request: CodeGenRequest,
        model_id: str,
        provider,
        prompt: str
    ) -> ModelResponse:
        """Execute model with retry logic"""
        
        @retry(
            stop=stop_after_attempt(3),
            wait=wait_exponential(multiplier=1, min=2, max=10),
            reraise=True
        )
        async def _execute():
            return await provider.generate_code(request, prompt)
        
        try:
            response = await _execute()
            
            # Validate syntax if successful
            if not response.error and response.generated_code:
                response.syntax_valid = await self._validate_syntax(
                    response.generated_code,
                    request.language
                )
                
                # Calculate quality metrics
                response.quality_metrics = await self._calculate_quality_metrics(
                    response.generated_code,
                    request.language
                )
            
            return response
            
        except Exception as e:
            logger.error(f"Model {model_id} failed after retries: {str(e)}")
            # Return error response
            model_config = self.model_registry.get_model(model_id)
            return ModelResponse(
                model_id=model_id,
                model_name=model_config.name if model_config else model_id,
                provider=model_config.provider.value if model_config else "unknown",
                generated_code="",
                confidence_score=0.0,
                execution_time_ms=0.0,
                tokens_used=0,
                prompt_tokens=0,
                completion_tokens=0,
                syntax_valid=False,
                error=str(e)
            )
    
    async def _validate_syntax(self, code: str, language: str) -> bool:
        """Validate code syntax"""
        # Basic syntax validation
        # In production, use language-specific parsers
        try:
            if language == "python":
                import ast
                ast.parse(code)
                return True
            elif language in ["javascript", "typescript"]:
                # Would use esprima or similar
                return len(code.strip()) > 0
            elif language == "go":
                # Would use go/parser
                return len(code.strip()) > 0
            else:
                return len(code.strip()) > 0
        except:
            return False
    
    async def _calculate_quality_metrics(
        self, 
        code: str, 
        language: str
    ) -> QualityMetrics:
        """Calculate basic quality metrics"""
        # Basic metrics calculation
        # In production, use proper static analysis tools
        
        lines = code.split('\n')
        lines_of_code = len([l for l in lines if l.strip() and not l.strip().startswith('#')])
        
        # Estimate complexity based on control flow keywords
        complexity_keywords = ['if', 'else', 'elif', 'for', 'while', 'match', 'case', 'switch']
        cyclomatic_complexity = 1 + sum(
            code.lower().count(keyword) for keyword in complexity_keywords
        )
        
        # Basic readability score (higher is better)
        avg_line_length = sum(len(l) for l in lines) / max(len(lines), 1)
        readability_score = max(0, 100 - (abs(avg_line_length - 80) / 2))
        
        # Documentation score based on comment ratio
        comment_lines = len([l for l in lines if l.strip().startswith('#') or '"""' in l])
        doc_score = min(100, (comment_lines / max(lines_of_code, 1)) * 200)
        
        return QualityMetrics(
            cyclomatic_complexity=cyclomatic_complexity,
            lines_of_code=lines_of_code,
            readability_score=readability_score,
            security_score=75.0,  # Would need proper security scanner
            documentation_score=doc_score
        )
    
    def _generate_consensus(
        self, 
        responses: List[ModelResponse]
    ) -> tuple[Optional[str], Optional[str]]:
        """Generate consensus code and determine best model"""
        
        # Simple strategy: choose highest confidence with valid syntax
        valid_responses = [r for r in responses if r.syntax_valid]
        
        if not valid_responses:
            valid_responses = responses
        
        # Sort by confidence score
        best_response = max(valid_responses, key=lambda r: r.confidence_score)
        
        return best_response.generated_code, best_response.model_id
    
    def _create_error_response(
        self, 
        request: CodeGenRequest, 
        error_message: str,
        elapsed_time: float
    ) -> CodeGenResponse:
        """Create error response"""
        return CodeGenResponse(
            request_id=request.request_id,
            status="failed",
            total_time_ms=elapsed_time * 1000,
            model_responses=[],
            metadata={"error": error_message}
        )
