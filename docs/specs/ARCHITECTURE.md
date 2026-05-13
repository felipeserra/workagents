# WorkAgents — Architecture

## Stack Decision

| Layer | Escolha | Motivo |
|-------|---------|--------|
| Backend | Go 1.23+ | Cold start ~5ms, binário único, nativo Vercel, 142k RPS |
| Frontend | React + Vite + shadcn/ui | Dashboard web interativo |
| Mobile | React Native + Expo | Código compartilhado com frontend web |
| Landing | React + Vite | Site público, SEO |
| Database | PostgreSQL (Neon serverless) | Serverless, compatível Vercel |
| ORM | Drizzle | Type-safe, leve |

## Deploy Architecture

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Landing     │     │  Frontend    │     │  Mobile      │
│  workagents  │     │  app.work    │     │  (Expo)      │
│  .com.br     │     │  agents.co   │     │              │
└──────┬───────┘     └──────┬───────┘     └──────┬───────┘
       │                    │                     │
       ▼                    ▼                     ▼
┌─────────────────────────────────────────────────────────┐
│              Vercel (Edge + Serverless)                   │
│  ┌────────────────────────────────────────────────────┐  │
│  │  apps/backend (Go Serverless Functions)             │  │
│  │  /api/companies /api/agents /api/tasks /api/auth   │  │
│  └────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────┐
│  Neon PostgreSQL            │
│  (Serverless, autoscaling)  │
└─────────────────────────────┘
```

## Vercel Configuration

```json
{
  "functions": {
    "apps/backend/api/**/*.go": {
      "runtime": "go@1.23",
      "maxDuration": 30,
      "memory": 512
    }
  }
}
```
