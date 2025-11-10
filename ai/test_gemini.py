"""
Test script for Gemini 2.5 Pro integration
Tests basic functionality of the Gemini provider
"""
import asyncio
import os
import sys
from pathlib import Path

# Add parent directory to path
sys.path.insert(0, str(Path(__file__).parent.parent))

from app.providers.gemini_provider import GeminiProvider
from app.config import settings


async def test_basic_generation():
    """Test basic code generation"""
    print("=" * 60)
    print("Test 1: Basic Code Generation")
    print("=" * 60)
    
    if not settings.google_api_key:
        print("❌ No Google API key found in environment")
        return False
    
    try:
        provider = GeminiProvider(
            model_id="gemini-2.5-pro",
            model_name="gemini-2.0-flash-exp",
            config={
                "temperature": 0.2,
                "top_p": 0.95,
                "top_k": 40,
                "max_output_tokens": 1024
            }
        )
        
        print("✅ Provider initialized successfully")
        print(f"Model: {provider.api_model_name}")
        print(f"LiteLLM model: {provider.litellm_model}")
        
        prompt = "Write a Python function that calculates the factorial of a number using recursion."
        
        print(f"\nPrompt: {prompt}")
        print("\nGenerating...")
        
        result = await provider.generate(prompt=prompt)
        
        print("\n" + "=" * 60)
        print("Generated Code:")
        print("=" * 60)
        print(result["content"])
        print("\n" + "=" * 60)
        print("Metadata:")
        print(f"  Model: {result['model']}")
        print(f"  Tokens used: {result['usage']['total_tokens']}")
        print(f"  Prompt tokens: {result['usage']['prompt_tokens']}")
        print(f"  Completion tokens: {result['usage']['completion_tokens']}")
        print(f"  Finish reason: {result['finish_reason']}")
        
        return True
        
    except Exception as e:
        print(f"❌ Error: {str(e)}")
        return False


async def test_json_mode():
    """Test JSON mode generation"""
    print("\n" + "=" * 60)
    print("Test 2: JSON Mode")
    print("=" * 60)
    
    if not settings.google_api_key:
        print("❌ No Google API key found in environment")
        return False
    
    try:
        provider = GeminiProvider(
            model_id="gemini-2.5-pro",
            model_name="gemini-2.0-flash-exp",
            config={
                "temperature": 0.2,
                "max_output_tokens": 512
            }
        )
        
        prompt = """Generate a JSON schema for a user object with the following fields:
        - id (string)
        - name (string)
        - email (string)
        - age (integer)
        - is_active (boolean)
        """
        
        print(f"Prompt: {prompt}")
        print("\nGenerating with JSON mode...")
        
        result = await provider.generate(prompt=prompt, json_mode=True)
        
        print("\n" + "=" * 60)
        print("Generated JSON:")
        print("=" * 60)
        print(result["content"])
        
        return True
        
    except Exception as e:
        print(f"❌ Error: {str(e)}")
        return False


async def test_system_instruction():
    """Test with system instruction"""
    print("\n" + "=" * 60)
    print("Test 3: System Instruction")
    print("=" * 60)
    
    if not settings.google_api_key:
        print("❌ No Google API key found in environment")
        return False
    
    try:
        provider = GeminiProvider(
            model_id="gemini-2.5-pro",
            model_name="gemini-2.0-flash-exp",
            config={
                "temperature": 0.3,
                "max_output_tokens": 1024
            }
        )
        
        system_instruction = """You are an expert Python developer who writes clean, 
        well-documented code following PEP 8 standards. Always include docstrings 
        and type hints."""
        
        prompt = "Create a function to merge two sorted lists into one sorted list."
        
        print(f"System Instruction: {system_instruction[:100]}...")
        print(f"\nPrompt: {prompt}")
        print("\nGenerating...")
        
        result = await provider.generate(
            prompt=prompt,
            system_instruction=system_instruction
        )
        
        print("\n" + "=" * 60)
        print("Generated Code:")
        print("=" * 60)
        print(result["content"])
        
        return True
        
    except Exception as e:
        print(f"❌ Error: {str(e)}")
        return False


async def test_health_check():
    """Test provider health check"""
    print("\n" + "=" * 60)
    print("Test 4: Health Check")
    print("=" * 60)
    
    if not settings.google_api_key:
        print("❌ No Google API key found in environment")
        return False
    
    try:
        provider = GeminiProvider(
            model_id="gemini-2.5-pro",
            model_name="gemini-2.0-flash-exp",
            config={}
        )
        
        print("Running health check...")
        is_healthy = await provider.health_check()
        
        if is_healthy:
            print("✅ Provider is healthy")
            return True
        else:
            print("❌ Provider health check failed")
            return False
            
    except Exception as e:
        print(f"❌ Error: {str(e)}")
        return False


async def test_provider_info():
    """Test getting provider information"""
    print("\n" + "=" * 60)
    print("Test 5: Provider Information")
    print("=" * 60)
    
    try:
        provider = GeminiProvider(
            model_id="gemini-2.5-pro",
            model_name="gemini-2.0-flash-exp",
            config={
                "temperature": 0.2,
                "top_p": 0.95,
                "top_k": 40
            }
        )
        
        info = provider.get_info()
        
        print("Provider Information:")
        print(f"  Model ID: {info['model_id']}")
        print(f"  Model Name: {info['model_name']}")
        print(f"  API Model Name: {info['api_model_name']}")
        print(f"  LiteLLM Model: {info['litellm_model']}")
        print(f"  Provider: {info['provider']}")
        print(f"\nFeatures:")
        for feature, enabled in info['features'].items():
            status = "✅" if enabled else "❌"
            print(f"  {status} {feature}: {enabled}")
        
        print(f"\nConfiguration:")
        for key, value in info['config'].items():
            print(f"  {key}: {value}")
        
        return True
        
    except Exception as e:
        print(f"❌ Error: {str(e)}")
        return False


async def main():
    """Run all tests"""
    print("\n" + "=" * 60)
    print("🚀 Gemini 2.5 Pro Integration Tests")
    print("=" * 60)
    
    # Check for API key
    if not settings.google_api_key or settings.google_api_key == "test-google-key":
        print("\n❌ No valid Google API key found!")
        print("\nPlease set your API key in .env:")
        print("  GOOGLE_API_KEY=your-actual-key-here")
        print("\nGet your key from:")
        print("  - https://makersuite.google.com/app/apikey")
        print("  - https://aistudio.google.com/app/apikey")
        return
    
    print(f"\n✅ Google API key found: {settings.google_api_key[:10]}...")
    
    # Run tests
    results = []
    
    results.append(("Basic Generation", await test_basic_generation()))
    results.append(("JSON Mode", await test_json_mode()))
    results.append(("System Instruction", await test_system_instruction()))
    results.append(("Health Check", await test_health_check()))
    results.append(("Provider Info", await test_provider_info()))
    
    # Summary
    print("\n" + "=" * 60)
    print("Test Summary")
    print("=" * 60)
    
    passed = sum(1 for _, result in results if result)
    total = len(results)
    
    for test_name, result in results:
        status = "✅ PASS" if result else "❌ FAIL"
        print(f"{status} - {test_name}")
    
    print(f"\n{passed}/{total} tests passed")
    
    if passed == total:
        print("\n🎉 All tests passed! Gemini integration is working correctly.")
    else:
        print(f"\n⚠️  {total - passed} test(s) failed. Check the errors above.")


if __name__ == "__main__":
    asyncio.run(main())
