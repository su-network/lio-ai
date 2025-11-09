# Lio AI

Multi-model AI code generation platform with support for:
- **Google Gemini 2.5 Pro** (Featured)
- OpenAI GPT-4/GPT-3.5
- Anthropic Claude 3
- Cohere Command R+

## Quick Start - Gemini 2.5 Pro

```bash
# 1. Setup Gemini
./scripts/setup_gemini.sh

# 2. Add your API key to .env
echo "GOOGLE_API_KEY=your-key" >> .env

# 3. Start services
make dev

# 4. Test it
curl -X POST http://localhost:8080/api/generate \
  -H "Content-Type: application/json" \
  -d '{"prompt":"Hello world","language":"python","selected_models":["gemini-2.5-pro"]}'
```

📚 **Full Documentation**: [Gemini Integration Guide](docs/GEMINI_INTEGRATION.md) | [Quick Reference](docs/GEMINI_QUICKSTART.md)

## Features

- 🚀 **Multi-Model Support**: Run multiple AI models in parallel
- 🎯 **Smart Model Selection**: Auto-select best model for your task
- 🔒 **Secure**: AES-256 encrypted API key storage
- ⚡ **Fast**: Optimized for performance with caching
- 🌐 **Multi-Language**: Support for 14+ programming languages
- 📊 **Monitoring**: Built-in metrics and health checks

Get your Gemini API key: https://makersuite.google.com/app/apikey