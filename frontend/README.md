# Lio AI Frontend

A modern, sleek Vue 3 web application for AI-powered code generation.

## Features

- **Document Upload**: Drag-and-drop file upload with support for PDF, TXT, JSON, YAML, and images
- **Code Generation**: Transform documents and prompts into functional code
- **AI Chat**: Interactive chat with multiple AI models
- **Document Management**: Organize and tag uploaded documents
- **Syntax Highlighting**: Code display with syntax highlighting
- **Responsive Design**: Works on desktop and mobile devices
- **Dark/Light Theme**: Modern UI with theme support

## Tech Stack

- **Vue 3** - Progressive JavaScript framework with Composition API
- **TypeScript** - Type-safe JavaScript
- **Vite** - Fast build tool and development server
- **Tailwind CSS** - Utility-first CSS framework
- **Pinia** - State management for Vue 3
- **Vue Router** - Client-side routing
- **Headless UI** - Unstyled, accessible components
- **Axios** - HTTP client for API calls
- **Lucide Vue** - Beautiful icons

## Getting Started

### Prerequisites

- Node.js 18+
- npm or yarn

### Installation

1. Install dependencies:
```bash
npm install
```

2. Start the development server:
```bash
npm run dev
```

3. Open [http://localhost:3000](http://localhost:3000) in your browser.

### Build for Production

```bash
npm run build
npm run preview
```

### Linting

```bash
npm run lint
```

### Formatting

```bash
npm run format
```

## Project Structure

```
src/
├── assets/          # Static assets and styles
├── components/      # Reusable Vue components
├── stores/          # Pinia stores for state management
├── types/           # TypeScript type definitions
├── utils/           # Utility functions
├── views/           # Page components
├── App.vue          # Root component
├── main.ts          # Application entry point
└── router/          # Vue Router configuration
```

## API Integration

The frontend communicates with the Lio AI backend services:

- **Go Gateway**: `http://localhost:8080` - Document management and routing
- **Python FastAPI**: `http://localhost:8000` - Code generation and AI inference

## Development

### Adding New Components

1. Create component in `src/components/`
2. Use Composition API with `<script setup>`
3. Follow Vue 3 and TypeScript best practices

### State Management

Use Pinia stores in `src/stores/` for global state:

```typescript
import { defineStore } from 'pinia'

export const useExampleStore = defineStore('example', () => {
  const state = ref(initialValue)

  const getter = computed(() => state.value)

  function action() {
    // mutate state
  }

  return { state, getter, action }
})
```

### Styling

Use Tailwind CSS classes directly in templates:

```vue
<template>
  <div class="bg-blue-500 text-white p-4 rounded-lg">
    Styled component
  </div>
</template>
```

## Contributing

1. Follow the existing code style
2. Use TypeScript for type safety
3. Write descriptive commit messages
4. Test your changes

## License

ISC