"""Main FastAPI application"""
from fastapi import FastAPI, HTTPException, BackgroundTasks
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from contextlib import asynccontextmanager
import time
import logging
import os
import httpx
from datetime import datetime

from app.config import settings
from app.models import (
    CodeGenRequest, CodeGenResponse, HealthCheck, ErrorResponse, ModelConfig
)
from app.services import ModelRegistry, PromptManager, CodeGenerationService

# Configure logging
logging.basicConfig(
    level=logging.INFO if not settings.debug else logging.DEBUG,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Global service instances
model_registry: ModelRegistry = None
prompt_manager: PromptManager = None
code_gen_service: CodeGenerationService = None
app_start_time = time.time()


async def sync_user_api_keys(user_id: str):
    """Fetch user's API keys from gateway and update environment"""
    try:
        gateway_url = os.getenv("GATEWAY_URL", "http://localhost:8080")
        async with httpx.AsyncClient() as client:
            response = await client.get(
                f"{gateway_url}/api/v1/api-keys",
                params={"user_id": user_id},
                timeout=5.0
            )
            
            if response.status_code == 200:
                data = response.json()
                keys = data.get("keys", [])
                
                # Update environment variables
                for key in keys:
                    provider = key.get("provider", "").lower()
                    # Fetch the actual decrypted key
                    key_response = await client.get(
                        f"{gateway_url}/api/v1/api-keys/{provider}",
                        params={"user_id": user_id},
                        timeout=5.0
                    )
                    
                    if key_response.status_code == 200:
                        key_data = key_response.json()
                        api_key = key_data.get("api_key")
                        
                        if api_key:
                            if provider == "openai":
                                os.environ["OPENAI_API_KEY"] = api_key
                                settings.openai_api_key = api_key
                            elif provider == "anthropic":
                                os.environ["ANTHROPIC_API_KEY"] = api_key
                                settings.anthropic_api_key = api_key
                            elif provider == "google":
                                os.environ["GOOGLE_API_KEY"] = api_key
                                settings.google_api_key = api_key
                            elif provider == "cohere":
                                os.environ["COHERE_API_KEY"] = api_key
                                settings.cohere_api_key = api_key
                            
                            logger.info(f"✓ Synced API key for {provider}")
                
                # Reload model registry to reinitialize providers
                model_registry.reload_config()
                logger.info("✓ Model registry reloaded with new API keys")
                
    except Exception as e:
        logger.error(f"Failed to sync API keys: {str(e)}")


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan manager"""
    global model_registry, prompt_manager, code_gen_service
    
    # Startup
    logger.info("Starting Lio AI Service...")
    
    try:
        # Initialize services
        model_registry = ModelRegistry(settings.models_config_path)
        prompt_manager = PromptManager(settings.prompts_config_path)
        code_gen_service = CodeGenerationService(model_registry, prompt_manager)
        
        logger.info(f"Loaded {len(model_registry.models)} models")
        logger.info("Service initialized successfully")
        
    except Exception as e:
        logger.error(f"Failed to initialize services: {str(e)}")
        raise
    
    yield
    
    # Shutdown
    logger.info("Shutting down Lio AI Service...")


# Create FastAPI app
app = FastAPI(
    title=settings.app_name,
    version=settings.app_version,
    description="Multi-Model Code Generation Service",
    lifespan=lifespan
)

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


# Exception handlers
@app.exception_handler(HTTPException)
async def http_exception_handler(request, exc):
    return JSONResponse(
        status_code=exc.status_code,
        content=ErrorResponse(
            error="HTTPException",
            message=exc.detail,
            details=str(exc)
        ).dict()
    )


@app.exception_handler(Exception)
async def general_exception_handler(request, exc):
    logger.error(f"Unhandled exception: {str(exc)}")
    return JSONResponse(
        status_code=500,
        content=ErrorResponse(
            error="InternalServerError",
            message="An unexpected error occurred",
            details=str(exc) if settings.debug else None
        ).dict()
    )


# Routes
@app.get("/", tags=["Root"])
async def root():
    """Root endpoint"""
    return {
        "service": settings.app_name,
        "version": settings.app_version,
        "status": "running",
        "docs": "/docs"
    }


@app.get("/health", response_model=HealthCheck, tags=["Health"])
async def health_check():
    """Health check endpoint"""
    try:
        # Check model health
        health_status = await model_registry.health_check_all()
        healthy_models = sum(1 for status in health_status.values() if status)
        
        return HealthCheck(
            status="healthy" if healthy_models > 0 else "unhealthy",
            version=settings.app_version,
            models_available=len(model_registry.models),
            models_healthy=healthy_models,
            uptime_seconds=time.time() - app_start_time
        )
    except Exception as e:
        logger.error(f"Health check failed: {str(e)}")
        raise HTTPException(status_code=500, detail="Health check failed")


@app.post("/api/v1/generate", response_model=CodeGenResponse, tags=["Code Generation"])
async def generate_code(request: CodeGenRequest):
    """
    Generate code using multiple AI models
    
    This endpoint orchestrates code generation across multiple LLM models,
    executes them in parallel, and returns aggregated results with quality metrics.
    """
    try:
        logger.info(f"Code generation request: {request.request_id}")
        
        # Generate code
        response = await code_gen_service.generate_code(request)
        
        logger.info(
            f"Request {request.request_id} completed: "
            f"{response.status}, {len(response.model_responses)} models"
        )
        
        return response
        
    except Exception as e:
        logger.error(f"Code generation failed: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=f"Code generation failed: {str(e)}"
        )


@app.get("/api/v1/models", tags=["Models"])
async def list_models(
    language: str = None,
    complexity: str = None,
    enabled_only: bool = True
):
    """
    List available models with optional filtering
    """
    try:
        from app.models import ComplexityLevel
        
        complexity_enum = None
        if complexity:
            complexity_enum = ComplexityLevel(complexity)
        
        models = model_registry.list_models(
            language=language,
            complexity=complexity_enum,
            enabled_only=enabled_only
        )
        
        return {
            "total": len(models),
            "available": len(models),
            "models": [model.dict() for model in models]
        }
        
    except Exception as e:
        logger.error(f"List models failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/api/v1/models/status", tags=["Models"])
async def get_models_status(user_id: str = None):
    """
    Get detailed status of all models including API key availability
    If user_id is provided, check against their stored API keys
    """
    try:
        # If user_id provided, fetch their API keys from gateway and update environment
        if user_id:
            await sync_user_api_keys(user_id)
        
        statuses = model_registry.get_all_models_status()
        
        available = sum(1 for s in statuses if s['status'] == 'available')
        no_api_key = sum(1 for s in statuses if s['status'] == 'no_api_key')
        disabled = sum(1 for s in statuses if s['status'] == 'disabled')
        
        return {
            "total_models": len(statuses),
            "available": available,
            "no_api_key": no_api_key,
            "disabled": disabled,
            "models": statuses
        }
        
    except Exception as e:
        logger.error(f"Get models status failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/api/v1/models/reload", tags=["Models"])
async def reload_models(user_id: str = None):
    """
    Reload model configuration and sync API keys
    This should be called after adding/updating API keys
    """
    try:
        # Sync API keys if user_id provided
        if user_id:
            await sync_user_api_keys(user_id)
        else:
            # Just reload from environment
            model_registry.reload_config()
        
        statuses = model_registry.get_all_models_status()
        available = sum(1 for s in statuses if s['status'] == 'available')
        
        return {
            "message": "Models reloaded successfully",
            "total_models": len(statuses),
            "available": available,
            "timestamp": datetime.utcnow().isoformat()
        }
    except Exception as e:
        logger.error(f"Reload models failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/api/v1/models/{model_id}", response_model=ModelConfig, tags=["Models"])
async def get_model(model_id: str):
    """Get details of a specific model"""
    model = model_registry.get_model(model_id)
    if not model:
        raise HTTPException(status_code=404, detail=f"Model {model_id} not found")
    return model


@app.post("/api/v1/models/{model_id}/health", tags=["Models"])
async def check_model_health(model_id: str):
    """Check health of a specific model"""
    provider = model_registry.get_provider(model_id)
    if not provider:
        raise HTTPException(status_code=404, detail=f"Model {model_id} not found")
    
    try:
        is_healthy = await provider.health_check()
        return {
            "model_id": model_id,
            "healthy": is_healthy,
            "timestamp": datetime.utcnow().isoformat()
        }
    except Exception as e:
        return {
            "model_id": model_id,
            "healthy": False,
            "error": str(e),
            "timestamp": datetime.utcnow().isoformat()
        }


@app.get("/api/v1/models/recommend", tags=["Models"])
@app.post("/api/v1/models/recommend", tags=["Models"])
async def recommend_models(
    language: str,
    complexity: str = "intermediate",
    framework: str = None,
    max_models: int = 3,
    strategy: str = "default"
):
    """
    Get recommended models for a specific request
    """
    try:
        from app.models import ComplexityLevel
        
        complexity_enum = ComplexityLevel(complexity)
        
        recommended = model_registry.select_models(
            language=language,
            complexity=complexity_enum,
            framework=framework,
            max_models=max_models,
            strategy=strategy
        )
        
        # Get full model configs
        models = [model_registry.get_model(mid) for mid in recommended]
        models = [m for m in models if m is not None]
        
        return {
            "language": language,
            "complexity": complexity,
            "framework": framework,
            "strategy": strategy,
            "recommended_count": len(models),
            "models": [model.dict() for model in models]
        }
        
    except Exception as e:
        logger.error(f"Model recommendation failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/api/v1/reload", tags=["Admin"])
async def reload_configuration(background_tasks: BackgroundTasks):
    """
    Reload configuration files (models and prompts)
    """
    try:
        def reload():
            model_registry.reload_config()
            prompt_manager.reload_config()
            logger.info("Configuration reloaded")
        
        background_tasks.add_task(reload)
        
        return {
            "message": "Configuration reload initiated",
            "timestamp": datetime.utcnow().isoformat()
        }
        
    except Exception as e:
        logger.error(f"Configuration reload failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/api/v1/stats", tags=["Statistics"])
async def get_statistics():
    """Get service statistics"""
    return {
        "uptime_seconds": time.time() - app_start_time,
        "total_models": len(model_registry.models),
        "enabled_models": len([m for m in model_registry.models.values() if m.enabled]),
        "providers": list(set(m.provider.value for m in model_registry.models.values())),
        "supported_languages": list(set(
            lang for m in model_registry.models.values() 
            for lang in m.capabilities.languages
        )),
        "timestamp": datetime.utcnow().isoformat()
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "app.main:app",
        host=settings.host,
        port=settings.port,
        reload=settings.debug
    )
