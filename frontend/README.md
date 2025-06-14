# Server Scheduler Frontend

This is the frontend application for the Server Scheduler system, built with Vue 3 and Element Plus.

## Project Setup

```bash
# Install dependencies
npm install

# Compile and hot-reload for development
npm run serve

# Compile and minify for production
npm run build

# Lint and fix files
npm run lint
```

## Features

- User authentication (login/register)
- Server management
- Reservation scheduling
- Responsive design
- Modern UI with Element Plus

## Development

The application uses:
- Vue 3 with Composition API
- Vuex for state management
- Vue Router for navigation
- Element Plus for UI components
- Axios for API communication

## API Integration

The frontend communicates with the backend API at `http://localhost:8080`. The development server is configured to proxy API requests to avoid CORS issues.

## Building for Production

To build the application for production:

1. Run `npm run build`
2. The built files will be in the `dist` directory
3. These files can be served by any static file server 