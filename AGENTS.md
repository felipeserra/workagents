# WorkAgents AGENTS.md

Guidance for human and AI contributors.

## 1. Purpose

WorkAgents is a control plane for autonomous AI companies — the infrastructure that AI agent workforces run on.

## 2. Read This First

1. `docs/specs/GOAL.md`
2. `docs/specs/PRODUCT.md`  
3. `docs/specs/SPEC.md`
4. `docs/specs/SPEC-impl.md`
5. `docs/specs/ARCHITECTURE.md`

## 3. Repo Map

- `apps/backend/` — Go API serverless (Vercel)
- `apps/frontend/` — React dashboard (board UI)
- `apps/mobile/` — React Native (Expo) mobile app
- `apps/landing/` — Site público (workagents.com.br)
- `packages/shared/` — Types, validators, constants
- `packages/db/` — Drizzle schema, migrations
- `packages/adapters/` — Agent adapters (Claude, Codex, OpenClaw, etc.)

## 4. Core Engineering Rules

1. **Company-scoped** — Every entity scoped to a company
2. **Contracts synchronized** — Schema → API → UI → Mobile
3. **Control-plane invariants** — Single-assignee task, atomic checkout, approval gates, budget hard-stop
4. **Spec alignment** — Implementation must match `docs/specs/SPEC-impl.md`
5. **Activity logging** — All mutations logged

## 5. Verification

```bash
pnpm build
pnpm test
pnpm lint
```
