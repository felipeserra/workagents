# WorkAgents — Platform

Control plane for autonomous AI companies. Backbone da economia autônoma.

## Stack

| Layer | Tecnologia | Deploy |
|-------|-----------|--------|
| Backend | Go 1.23+ | Vercel Serverless Functions |
| Frontend | React + Vite + shadcn/ui | Vercel |
| Mobile | React Native + Expo | EAS / App Store |
| Landing | React + Vite | Vercel |
| Database | PostgreSQL (Neon) | Neon |
| ORM | Drizzle | — |
| Monorepo | Turborepo + pnpm | — |

## Estrutura

```
workagents/
├── apps/
│   ├── backend/        ← Go API (serverless)
│   ├── frontend/       ← React dashboard web
│   ├── mobile/         ← React Native (Expo)
│   └── landing/        ← Site público
├── packages/
│   ├── shared/         ← Types, validators, schemas
│   ├── db/             ← Drizzle schema + migrations
│   └── adapters/       ← Adaptadores de agentes
├── docs/
│   ├── specs/          ← Spec Kit (especificações)
│   └── plans/          ← Planos de implementação
├── docker/             ← Deploy self-hosted
└── .github/workflows/  ← CI/CD
```

## Desenvolvimento

```bash
pnpm install
pnpm dev              # Backend + frontend
pnpm landing:dev      # Landing page
pnpm mobile:start     # Expo
```

## Deploy

```bash
# Landing
cd apps/landing && vercel deploy --prod

# Frontend
cd apps/frontend && vercel deploy --prod

# Backend (Go)
cd apps/backend && vercel deploy --prod
```

## Licença

MIT
