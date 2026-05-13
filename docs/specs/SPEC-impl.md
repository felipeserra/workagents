# WorkAgents — SPEC Implementation

Status: V1 build contract
Date: 2026-05-13

## 1. V1 Outcomes

1. Board cria company e define goals
2. Board cria e gerencia agentes em org tree
3. Agentes recebem e executam tarefas via heartbeat
4. Todo trabalho rastreado com audit visibility
5. Custos reportados com budget limits
6. Board pode intervir em qualquer lugar

## 2. Modules

### Module A: Auth & Companies
- [ ] POST /api/auth/login — login board
- [ ] POST /api/companies — criar company
- [ ] GET /api/companies — listar companies
- [ ] GET /api/companies/:id — detalhes
- [ ] PATCH /api/companies/:id — atualizar
- [ ] DELETE /api/companies/:id — deletar

### Module B: Agents
- [ ] POST /api/agents — criar agente
- [ ] GET /api/agents — listar agentes
- [ ] GET /api/agents/:id — detalhes
- [ ] PATCH /api/agents/:id — atualizar
- [ ] DELETE /api/agents/:id — deletar
- [ ] POST /api/agents/:id/pause — pausar
- [ ] POST /api/agents/:id/resume — retomar

### Module C: Tasks
- [ ] POST /api/tasks — criar tarefa
- [ ] GET /api/tasks — listar tarefas
- [ ] PATCH /api/tasks/:id — atualizar
- [ ] POST /api/tasks/:id/checkout — checkout atômico
- [ ] POST /api/tasks/:id/complete — completar
- [ ] POST /api/tasks/:id/comment — comentar

### Module D: Heartbeats
- [ ] POST /api/heartbeats — trigger heartbeat
- [ ] GET /api/heartbeats — listar heartbeats
- [ ] GET /api/heartbeats/:id/logs — logs do heartbeat

### Module E: Budgets
- [ ] POST /api/budgets — definir orçamento
- [ ] GET /api/budgets — consultar orçamentos
- [ ] GET /api/budgets/usage — uso atual

### Module F: Activity Log
- [ ] GET /api/activity — feed de atividade

## 3. Tech Specs

**Backend (Go):**
- Standard library `net/http` + `chi` router
- Drizzle ORM via Go (sqlc ou raw SQL)
- JWT para auth board
- API keys hashed para agentes

**Frontend (React):**
- Vite + React 19 + TypeScript
- shadcn/ui components
- React Query para data fetching
- React Router v7 para rotas

**Mobile (React Native + Expo):**
- Expo SDK 52+
- React Navigation
- expo-router

## 4. Database Schema (V1)

```sql
companies (id, name, goal, budget, created_at, updated_at)
agents (id, company_id, name, role, adapter_type, adapter_config, reports_to, capabilities, status, created_at)
tasks (id, company_id, agent_id, title, description, status, priority, budget_spent, created_at, updated_at)
task_comments (id, task_id, agent_id, content, created_at)
heartbeats (id, agent_id, status, started_at, completed_at, logs)
budgets (id, company_id, agent_id, period_start, period_end, limit_amount, spent_amount)
activity_logs (id, company_id, actor_id, action, target_type, target_id, metadata, created_at)
```
