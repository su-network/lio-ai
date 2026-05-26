# Lio AI: A Plug-and-Play Multi-Service Platform for Document-Aware AI Chat and Code Generation

**Author(s):** _[Your Name / Team Name]_  
**Date:** 2026-05-26  
**Repository:** https://github.com/su-network/lio-ai (default branch: `master`)  
**Artifact Type:** Research Experience / Technical Project Report

## Abstract

This report documents the design and implementation of **Lio AI**, an end-to-end platform that combines document management, conversational AI, and code generation in a single developer workflow. The current implementation integrates a Vue 3 frontend, a Go API gateway, and a Python FastAPI model orchestration service. The system emphasizes “plug-and-play” operation through Makefile-driven local development and Docker-based quickstart, while providing authenticated chat, document CRUD, model selection, usage tracking, and proxy-based code-generation endpoints.

## 1. Problem Statement and Motivation

Modern AI-assisted development workflows are often fragmented: one tool for document handling, another for chat, and another for model-backed code generation. This fragmentation increases integration cost and complicates reproducibility.

Lio AI addresses this by providing a unified stack where users can:
- upload/manage documents,
- run authenticated chat sessions with persisted message history,
- invoke AI-backed code generation and model endpoints through a consistent gateway.

The project motivation, as reflected in repository description (“upload, chat and much more”), is to reduce setup friction and provide an extensible base for AI-enabled application workflows.

## 2. System Overview

The current implementation supports:
- **Frontend UX (Vue + TypeScript):** chat interface, model/status views, settings/API-key interactions, and document/code-generation workflows (`frontend/src`).
- **Gateway/API layer (Go + Gin):** auth, document/chat/usage/system routes, plus protected proxying to AI service (`joles/cmd/server/main.go`).
- **AI orchestration service (Python + FastAPI):** model registry, model status/reload, code generation, chat completion, and health/statistics endpoints (`ai/app/main.py`).
- **Persistence and accounting:** SQLite-backed tables for users, documents, chats/messages, usage metrics, quotas, and provider API keys (`joles/internal/db/database.go`).

## 3. Architecture

```mermaid
flowchart LR
    U[User Browser] --> FE[Vue 3 Frontend<br/>frontend/src]
    FE -->|HTTP + cookies + CSRF header| GW[Go Gateway (Gin)<br/>joles/cmd/server]
    GW --> DB[(SQLite<br/>data/lio.db)]
    GW -->|Proxy /api/v1/codegen/*,<br/>/api/v1/models/*, /api/v1/stats| AI[Python FastAPI AI Service<br/>ai/app/main.py]
    AI --> MC[Model Registry + Providers<br/>ai/app/services]
    MC --> LP[LiteLLM / Gemini / Ollama Providers]

    GW --> SEC[JWT + CSRF + Rate Limiting<br/>middleware]
    FE --> TSU[TypeScript API client + stores<br/>frontend/src/services, stores]
```

### 3.1 Data/Control Flow Summary
1. Frontend sends requests to gateway endpoints (auth/doc/chat/usage/system).
2. Gateway enforces middleware (JWT, CSRF, rate limit) and handles local persistence where applicable.
3. AI-oriented calls are proxied to the FastAPI service.
4. FastAPI selects/validates models and executes generation/chat operations through provider adapters.
5. Responses are returned to frontend for rendering and interaction continuity.

## 4. Implementation Highlights by Language

### 4.1 Vue (UI and Interaction Layer)
- Composition API + Pinia store pattern (`frontend/src/main.ts`, `frontend/src/stores/*`).
- Auth-gated routing with login/register/profile flows (`frontend/src/router/index.ts`).
- Chat state management with conversation UUID support and model selection (`frontend/src/stores/chat.ts`).
- Frontend README documents feature scope including document upload, chat, and code generation (`frontend/README.md`).

### 4.2 Go (Gateway, Business APIs, Security Middleware)
- Gin router composition for `/api/v1/auth`, `/documents`, `/chats`, `/usage`, `/system`, `/api-keys` (`joles/cmd/server/main.go`).
- Proxy groups to AI service for `/api/v1/codegen/*`, `/api/v1/models/*`, and `/api/v1/stats`.
- Security middleware includes JWT auth and CSRF token verification (`joles/internal/middleware/auth_middleware.go`, `joles/internal/middleware/csrf.go`).
- SQLite migrations define core operational schema (`joles/internal/db/database.go`).

### 4.3 Python (Model Orchestration and AI Serving)
- FastAPI application with lifecycle initialization for model registry and prompt manager (`ai/app/main.py`).
- Multi-model code generation orchestration with async parallel execution and retry policies (`ai/app/services/code_generation.py`).
- Provider abstraction and API-key-aware model availability (`ai/app/services/model_registry.py`).
- Config-driven prompts/models (`ai/config/prompts.yaml`, `ai/config/models.yaml`).

### 4.4 TypeScript (Client Utilities and API Integration)
- Centralized API client with typed request/response structures and interceptors (`frontend/src/services/api.ts`).
- CSRF header injection for state-changing requests and cookie-based auth integration.
- Utility-level response obfuscation/deobfuscation helpers (`frontend/src/utils/encryption.ts`), currently documented in code as lightweight obfuscation (not strong encryption).

## 5. Evaluation and Validation

### 5.1 Testing Strategy
- Makefile defines primary validation paths: `make test`, `make build`, `make lint`, `make vet`, `make test-coverage`.
- In this environment:
  - `make build` succeeds for Go gateway.
  - `make test` currently fails due to pre-existing compile/test issues under `joles/tests`.
  - `make lint` expects `golangci-lint` installed.

This indicates the repository has an existing validation scaffold, though CI/local setup may require dependency/tool alignment.

### 5.2 Performance Considerations
- AI service executes model calls concurrently for code generation (`asyncio.gather` with timeout/retry).
- Model selection attempts to balance capability, cost, and latency factors (`ModelRegistry.select_models`).
- Gateway includes rate-limiting middleware and usage metrics/cost tracking endpoints.

### 5.3 Security and Privacy Considerations
- JWT-based authentication with protected API groups.
- CSRF defenses for state-changing requests via token cookie/header match.
- Provider API keys are handled via dedicated backend routes and persisted in encrypted form at rest (`provider_api_keys` table schema indicates encrypted storage field).
- Frontend utility obfuscation should not be interpreted as cryptographic protection; production-grade cryptography should remain server-side.

## 6. Limitations and Future Work

Current limitations (from implementation and docs):
- Validation status is not fully green in baseline local testing (`make test` failures in current state).
- Some security/performance capabilities are scaffolded but could be hardened further (e.g., stronger client-side secret handling assumptions, broader test coverage).
- Documentation is currently sparse at repository root; architecture and operational docs can be expanded.

Suggested future work:
- Stabilize and expand automated tests across Go/Python/frontend boundaries.
- Add structured benchmarks (latency/cost/quality comparisons across models).
- Introduce richer observability dashboards and SLO-based alerting.
- Extend RAG/document ingestion pipelines with stronger provenance and evaluation metrics.

## 7. Reproducibility

The following steps are directly derived from repository docs and Makefile targets.

### 7.1 Quickstart via Docker (root README)
```bash
docker build -t lio-ai .
docker run --rm -p 8080:8080 lio-ai
```

### 7.2 Local Multi-Service Development (Makefile)
```bash
make deps        # install/check Go + Python + frontend dependencies
make build       # build Go gateway
make ai-dev      # start FastAPI AI service in background
make frontend-dev
make dev         # orchestrated dev mode (Go + Python + frontend)
make status      # service status
make stop-all    # stop everything
```

### 7.3 Validation Commands
```bash
make test
make test-coverage
make vet
make lint
```

## 8. References and Related Materials

1. **Repository root README**: `/README.md`  
2. **Frontend documentation**: `/frontend/README.md`  
3. **Gateway entrypoint and routing**: `/joles/cmd/server/main.go`  
4. **Gateway system/metrics handler**: `/joles/internal/handlers/system_handler.go`  
5. **Database schema/migrations**: `/joles/internal/db/database.go`  
6. **Security middleware**: `/joles/internal/middleware/auth_middleware.go`, `/joles/internal/middleware/csrf.go`  
7. **FastAPI app**: `/ai/app/main.py`  
8. **Model/codegen services**: `/ai/app/services/model_registry.py`, `/ai/app/services/code_generation.py`  
9. **Model/prompt configs**: `/ai/config/models.yaml`, `/ai/config/prompts.yaml`

---

_End of report._
