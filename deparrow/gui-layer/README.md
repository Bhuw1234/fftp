# DEparrow GUI Layer

Modern web interface for the DEparrow distributed compute platform. Built with React, TypeScript, and Tailwind CSS.

## Features

- **Dashboard**: Real-time platform overview with stats and monitoring
- **Job Management**: Submit, monitor, and manage compute jobs
- **Node Monitoring**: View compute node status and resource utilization
- **Wallet System**: Credit-based payment system with transaction history
- **User Authentication**: Secure login and account management
- **Responsive Design**: Works on desktop and mobile devices

## Tech Stack

- **Frontend**: React 18 with TypeScript
- **Styling**: Tailwind CSS with custom design system
- **State Management**: React Query for server state
- **Routing**: React Router v6
- **HTTP Client**: Axios with interceptors
- **UI Components**: Custom components with Lucide React icons
- **Build Tool**: Vite for fast development and building
- **Notifications**: React Hot Toast

## Project Structure

```
src/
├── api/              # API client and service definitions
├── components/       # Reusable UI components
├── contexts/        # React contexts (Auth, Theme, etc.)
├── pages/           # Page components
│   ├── Dashboard.tsx
│   ├── Jobs.tsx
│   ├── Nodes.tsx
│   ├── Wallet.tsx
│   ├── Settings.tsx
│   └── Login.tsx
├── types/           # TypeScript type definitions
├── utils/           # Utility functions
├── App.tsx          # Main app component
├── main.tsx         # Entry point
└── index.css        # Global styles
```

## Getting Started

### Prerequisites

- Node.js 18+ and npm/yarn/pnpm

### Installation

```bash
cd deparrow/gui-layer
npm install
```

### Development

```bash
npm run dev
```

The development server will start at `http://localhost:3000`.

### Building for Production

```bash
npm run build
```

The built files will be in the `dist` directory.

### Linting

```bash
npm run lint
```

## API Integration

The GUI connects to the DEparrow Meta-OS bootstrap server (default: `http://localhost:8080`). Configure the API URL in:

1. `vite.config.ts` - Development proxy
2. Environment variables for production

### Environment Variables

Create a `.env` file:

```env
VITE_API_URL=http://localhost:8080/api
```

## Key Components

### Authentication
- JWT-based authentication with token storage
- Protected routes and API interceptors
- User context for global state management

### Job Submission
- Multi-step job creation wizard
- Template support for common job types
- Real-time job status updates
- Log streaming and result download

### Node Management
- Real-time node status monitoring
- Resource utilization charts
- Region-based node filtering
- Node health checks

### Wallet System
- Credit balance display
- Deposit/withdrawal functionality
- Transaction history
- Job cost estimation

## Design System

### Colors
- Primary: Blue (#3B82F6)
- Success: Green (#10B981)
- Warning: Yellow (#F59E0B)
- Error: Red (#EF4444)
- Background: Gray (#F9FAFB)

### Typography
- Font: Inter
- Base size: 16px
- Scale: 0.75rem → 3rem

### Spacing
- Base unit: 0.25rem (4px)
- Scale: 0.25, 0.5, 1, 1.5, 2, 2.5, 3, 4, 5, 6, 8, 10, 12, 14, 16, 20, 24, 28, 32, 36, 40, 44, 48, 52, 56, 60, 64

## Development Guidelines

### Component Structure
- Use functional components with TypeScript
- Implement proper error boundaries
- Follow React hooks best practices
- Use React Query for data fetching

### State Management
- Local state: `useState`, `useReducer`
- Server state: React Query
- Global state: React Context
- Form state: React Hook Form

### Styling
- Use Tailwind CSS utility classes
- Create reusable component variants
- Follow mobile-first responsive design
- Use CSS variables for theming

### Testing
- Unit tests with Jest and React Testing Library
- Component testing with Storybook
- E2E tests with Cypress

## Deployment

### Docker

```dockerfile
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

### Static Hosting
- Netlify, Vercel, GitHub Pages
- S3 + CloudFront
- Any static file server

## API Endpoints

The GUI expects the following API endpoints:

### Authentication
- `POST /api/auth/login`
- `POST /api/auth/register`
- `POST /api/auth/logout`
- `GET /api/auth/me`

### Jobs
- `GET /api/jobs`
- `GET /api/jobs/:id`
- `POST /api/jobs`
- `POST /api/jobs/:id/cancel`
- `GET /api/jobs/:id/logs`
- `GET /api/jobs/:id/results`

### Nodes
- `GET /api/nodes`
- `GET /api/nodes/:id`
- `GET /api/nodes/stats`

### Wallet
- `GET /api/wallet/balance`
- `GET /api/wallet/transactions`
- `POST /api/wallet/deposit`
- `POST /api/wallet/withdraw`

### System
- `GET /api/health`
- `GET /api/stats`
- `GET /api/config`

## License

Part of the DEparrow distributed compute platform. See main project LICENSE.