# Frontend Test Suite

This directory contains comprehensive tests for the ServerScheduler frontend application built with Vue.js 3, Element Plus, and Vuex.

## Test Structure

```
tests/
├── setup.js                 # Global test configuration and mocks
├── unit/
│   ├── components/          # Component tests
│   │   └── App.spec.js     # Main App component tests
│   ├── store/              # Vuex store tests
│   │   └── auth.spec.js    # Authentication store module tests
│   ├── utils/              # Utility function tests
│   │   └── api.spec.js     # API client configuration tests
│   └── views/              # View component tests
│       ├── Calendar.spec.js     # Calendar view tests
│       ├── Login.spec.js        # Login view tests
│       ├── Reservations.spec.js # Reservations view tests
│       └── Servers.spec.js      # Servers view tests
└── README.md               # This file
```

## Testing Framework

- **Jest**: JavaScript testing framework
- **Vue Test Utils**: Official testing utilities for Vue.js
- **axios-mock-adapter**: HTTP request mocking
- **jsdom**: DOM simulation for testing

## Test Coverage

### Store Tests (`store/`)
- **auth.spec.js**: Authentication store module
  - Getters (isAuthenticated, user, token)
  - Mutations (SET_TOKEN, SET_USER, CLEAR_AUTH)
  - Actions (login, register, logout, fetchUser, initializeAuth)
  - Error handling and edge cases

### Component Tests (`components/`)
- **App.spec.js**: Main application component
  - Layout rendering for authenticated/unauthenticated users
  - Navigation menu functionality
  - User dropdown and logout
  - Change password dialog
  - Responsive design features
  - Theme and styling

### View Tests (`views/`)

#### Login.spec.js
- Form rendering and validation
- Login submission and error handling
- Loading states
- Navigation for authenticated users
- Form validation rules

#### Servers.spec.js
- Server list display and management
- CRUD operations (Create, Read, Update, Delete)
- Password visibility and clipboard functionality
- Role-based access control
- Form validation and error handling
- Status display and filtering

#### Reservations.spec.js
- Reservation list display and management
- CRUD operations for reservations
- Credential management and copying
- Status display and formatting
- Date formatting and validation
- Server selection and filtering
- Role-based access control

#### Calendar.spec.js
- Calendar component rendering
- Event processing and display
- Server filtering functionality
- Date formatting and conversion
- Event details dialog
- Status styling and CSS classes
- Error handling and edge cases

### Utility Tests (`utils/`)
- **api.spec.js**: API client configuration
  - Base configuration (URL, timeout, headers)
  - Request interceptors (authentication headers)
  - Response interceptors (error handling)
  - HTTP methods (GET, POST, PUT, DELETE, PATCH)
  - Error response handling
  - Content type handling

## Running Tests

### All Tests
```bash
npm test
# or
npm run test:unit
```

### Watch Mode
```bash
npm run test:unit:watch
```

### Coverage Report
```bash
npm run test:unit:coverage
```

### Using Makefile (from project root)
```bash
make frontend-test          # Run tests once
make frontend-test-watch    # Run in watch mode
make frontend-test-coverage # Generate coverage report
make test-all              # Run both backend and frontend tests
```

## Test Configuration

### Jest Configuration (`package.json`)
```json
{
  "jest": {
    "testEnvironment": "jsdom",
    "moduleFileExtensions": ["js", "json", "vue"],
    "transform": {
      "^.+\\.vue$": "@vue/vue3-jest",
      "^.+\\.js$": "babel-jest"
    },
    "testMatch": [
      "**/tests/unit/**/*.spec.js",
      "**/tests/unit/**/*.test.js"
    ],
    "collectCoverageFrom": [
      "src/**/*.{js,vue}",
      "!src/main.js",
      "!**/node_modules/**"
    ],
    "setupFilesAfterEnv": ["<rootDir>/tests/setup.js"]
  }
}
```

### Global Setup (`setup.js`)
- Element Plus component mocks
- Vue Router mocks
- Vuex store mocks
- localStorage mocks
- Console method mocks
- ResizeObserver and IntersectionObserver mocks

## Mocking Strategy

### API Mocking
Uses `axios-mock-adapter` to mock HTTP requests:
```javascript
const mockAxios = new MockAdapter(apiClient)
mockAxios.onGet('/api/servers').reply(200, mockData)
```

### Component Mocking
Element Plus components are stubbed for testing:
```javascript
stubs: {
  'el-table': { template: '<div class="el-table"><slot /></div>' },
  'el-button': { 
    template: '<button @click="$emit(\'click\')"><slot /></button>',
    emits: ['click']
  }
}
```

### Store Mocking
Vuex store modules are mocked with Jest functions:
```javascript
const mockStore = createStore({
  modules: {
    auth: {
      namespaced: true,
      actions: { login: jest.fn() },
      getters: { isAuthenticated: () => true }
    }
  }
})
```

## Test Patterns

### Component Testing Pattern
```javascript
describe('ComponentName.vue', () => {
  let wrapper
  let store
  let router

  beforeEach(() => {
    // Setup store, router, and other dependencies
    wrapper = mount(Component, {
      global: { plugins: [store, router] }
    })
  })

  afterEach(() => {
    wrapper.unmount()
  })

  describe('Feature Group', () => {
    it('should do something specific', async () => {
      // Arrange, Act, Assert
    })
  })
})
```

### Async Testing
```javascript
it('should handle async operations', async () => {
  mockAxios.onPost('/api/endpoint').reply(200, responseData)
  
  await wrapper.vm.someAsyncMethod()
  
  expect(wrapper.vm.someProperty).toBe(expectedValue)
})
```

### Error Handling Testing
```javascript
it('should handle errors gracefully', async () => {
  mockAxios.onGet('/api/endpoint').reply(500, { error: 'Server error' })
  
  await wrapper.vm.fetchData()
  
  expect(global.ElMessage.error).toHaveBeenCalledWith('Failed to load data')
})
```

## Coverage Goals

- **Statements**: > 80%
- **Branches**: > 75%
- **Functions**: > 80%
- **Lines**: > 80%

## Best Practices

1. **Descriptive Test Names**: Use clear, descriptive test names that explain what is being tested
2. **Arrange-Act-Assert**: Structure tests with clear setup, execution, and assertion phases
3. **Mock External Dependencies**: Mock API calls, external libraries, and complex components
4. **Test User Interactions**: Test clicks, form submissions, and other user interactions
5. **Test Error States**: Include tests for error conditions and edge cases
6. **Async Testing**: Properly handle asynchronous operations with async/await
7. **Clean Up**: Unmount components and restore mocks after each test

## CI/CD Integration

Tests are automatically run in the CI/CD pipeline:
- On every push to main/develop branches
- On every pull request
- Coverage reports are uploaded to Codecov
- Tests must pass before deployment

## Debugging Tests

### Running Single Test File
```bash
npx jest tests/unit/views/Login.spec.js
```

### Running Tests with Debug Output
```bash
npx jest --verbose tests/unit/views/Login.spec.js
```

### Debug Mode
```bash
node --inspect-brk node_modules/.bin/jest --runInBand tests/unit/views/Login.spec.js
```

## Contributing

When adding new features:
1. Write tests for new components/functionality
2. Ensure tests pass locally before committing
3. Maintain or improve test coverage
4. Follow existing test patterns and conventions
5. Update this documentation if adding new test categories 