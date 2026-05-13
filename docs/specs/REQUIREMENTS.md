# WorkAgents — Requisitos Detalhados V1

> Baseado no site workagents.com.br, Spec Kit docs e legado Paperclip SPEC.md

---

## 1. Empresas (Multi-company)

### User Stories
- **US-01** — Board deve criar empresa informando nome e goal
- **US-02** — Board deve listar todas as empresas
- **US-03** — Board deve editar dados da empresa
- **US-04** — Board deve excluir empresa (soft delete)
- **US-05** — Cada empresa é isolada: agentes e tarefas não cruzam

### Regras de Negócio
| # | Regra |
|---|-------|
| RN01 | Toda entidade é scoped por company_id |
| RN02 | Board tem acesso irrestrito a todas as empresas da instância |
| RN03 | Soft delete: companies.ativo = false |

### Endpoints
| Método | Rota | Descrição |
|--------|------|-----------|
| POST | /api/companies | Criar empresa |
| GET | /api/companies | Listar empresas |
| GET | /api/companies/:id | Detalhes da empresa |
| PATCH | /api/companies/:id | Atualizar empresa |
| DELETE | /api/companies/:id | Deletar (soft) |

---

## 2. Agentes (Org Structure)

### User Stories
- **US-06** — Board deve criar agente vinculado a uma empresa
- **US-07** — Agente tem: nome, role, adapter_type, adapter_config, reports_to
- **US-08** — Board deve montar organograma hierárquico (árvore)
- **US-09** — Board pode pausar/retomar qualquer agente
- **US-10** — Board pode excluir agente

### Regras de Negócio
| # | Regra |
|---|-------|
| RN04 | Org graph é árvore estrita (reports_to nullable = CEO) |
| RN05 | Cada agente publica capabilities (ajuda discovery entre agentes) |
| RN06 | Pause: sinal graceful + stop heartbeats futuros. Resume: reativa heartbeats |
| RN07 | Adapter types built-in V1: process, http |
| RN08 | Adapter config é blob opaco (schema definido pelo adapter) |
| RN09 | Visibilidade total: todos os agentes veem toda a org |

### Endpoints
| Método | Rota | Descrição |
|--------|------|-----------|
| POST | /api/agents | Criar agente |
| GET | /api/agents | Listar agentes (filtro: company_id) |
| GET | /api/agents/:id | Detalhes |
| PATCH | /api/agents/:id | Atualizar |
| DELETE | /api/agents/:id | Excluir |
| POST | /api/agents/:id/pause | Pausar |
| POST | /api/agents/:id/resume | Retomar |

---

## 3. Tarefas (Heartbeat Execution)

### User Stories
- **US-11** — Board cria tarefa e atribui a um agente
- **US-12** — Agente faz checkout atômico (ninguém mais pega)
- **US-13** — Agente completa tarefa com resultado
- **US-14** — Agente comenta em tarefas
- **US-15** — Toda tarefa tem: título, descrição, status, prioridade, assignee

### Regras de Negócio
| # | Regra |
|---|-------|
| RN10 | Single assignee por tarefa |
| RN11 | Checkout atômico: transição para in_progress bloqueia outros |
| RN12 | Tarefa sem assignee fica em backlog |
| RN13 | Histórico de comentários é imutável (append-only) |
| RN14 | Tarefas podem ser hierárquicas (pai-filho) via parent_id |

### Endpoints
| Método | Rota | Descrição |
|--------|------|-----------|
| POST | /api/tasks | Criar tarefa |
| GET | /api/tasks | Listar (filtros: company_id, agent_id, status) |
| GET | /api/tasks/:id | Detalhes + comentários |
| PATCH | /api/tasks/:id | Atualizar |
| POST | /api/tasks/:id/checkout | Checkout atômico |
| POST | /api/tasks/:id/complete | Completar |
| POST | /api/tasks/:id/comment | Adicionar comentário |

---

## 4. Heartbeats

### User Stories
- **US-16** — Sistema dispara heartbeat para agente no schedule configurado
- **US-17** - Board pode trigger heartbeat manual
- **US-18** — Heartbeat registra status (running, completed, failed)
- **US-19** — Modos: run command (monitorado) ou fire-and-forget

### Regras de Negócio
| # | Regra |
|---|-------|
| RN15 | Heartbeat invoca adapter (process | http) |
| RN16 | Adapter contract: invoke(), status(), cancel() |
| RN17 | Pause → sinal graceful + stop heartbeats futuros |
| RN18 | Context delivery: thin ping (só wake) ou fat payload (com contexto) |

### Endpoints
| Método | Rota | Descrição |
|--------|------|-----------|
| POST | /api/heartbeats | Trigger heartbeat |
| GET | /api/heartbeats | Listar (filtro: agent_id) |
| GET | /api/heartbeats/:id/logs | Logs do heartbeat |

---

## 5. Budgets (Cost Tracking)

### User Stories
- **US-20** — Board define orçamento por empresa/agente
- **US-21** — Sistema alerta ao atingir 80% do budget
- **US-22** — Hard-stop: auto-pause agente ao atingir 100%
- **US-23** — Board pode sobrescrever qualquer budget

### Regras de Negócio
| # | Regra |
|---|-------|
| RN19 | Budget tracking em tokens e dólares |
| RN20 | Três tiers: visibility → soft alert → hard ceiling |
| RN21 | Apenas Board pode modificar budgets |
| RN22 | Período de budget = billing cycle |

### Endpoints
| Método | Rota | Descrição |
|--------|------|-----------|
| POST | /api/budgets | Criar orçamento |
| GET | /api/budgets | Listar orçamentos |
| GET | /api/budgets/usage | Uso atual (tokens + $) |
| PATCH | /api/budgets/:id | Atualizar limite |
| POST | /api/budgets/:id/override | Board override |

---

## 6. Activity Log (Full Audit)

### User Stories
- **US-24** — Toda ação mutante é logada
- **US-25** — Board consulta feed de atividade por empresa
- **US-26** — Log contém: actor, action, target_type, target_id, metadata

### Regras de Negócio
| # | Regra |
|---|-------|
| RN23 | Activity log é append-only, imutável |
| RN24 | Toda mutation escreve log automaticamente |
| RN25 | Board vê tudo. Agentes veem apenas sua empresa |

### Endpoints
| Método | Rota | Descrição |
|--------|------|-----------|
| GET | /api/activity | Feed de atividade (filtro: company_id) |

---

## 7. Board Governance

### User Stories
- **US-27** — Board faz login (JWT)
- **US-28** — Approval gate: criar novo agente requer aprovação
- **US-29** — Approval gate: breakdown estratégico do CEO requer aprovação
- **US-30** — Board pode override qualquer decisão de agente

### Regras de Negócio
| # | Regra |
|---|-------|
| RN26 | V1: single human board operator |
| RN27 | Board autentica via email+senha ou OAuth Google |
| RN28 | Approval gates são async: Board recebe notificação, aprova/rejeita |
| RN29 | Board override é logado no activity log |

### Endpoints
| Método | Rota | Descrição |
|--------|------|-----------|
| POST | /api/auth/login | Login board |
| POST | /api/auth/register | Registrar board |
| POST | /api/approvals | Criar approval request |
| GET | /api/approvals | Listar pendentes |
| POST | /api/approvals/:id/approve | Aprovar |
| POST | /api/approvals/:id/reject | Rejeitar |

---

## Database Schema V1

```sql
-- Companies
companies (
  id UUID PK,
  name TEXT NOT NULL,
  goal TEXT,
  budget_limit NUMERIC(12,2) DEFAULT 0,
  active BOOLEAN DEFAULT true,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
)

-- Agents
agents (
  id UUID PK,
  company_id UUID FK → companies,
  name TEXT NOT NULL,
  role TEXT NOT NULL,
  adapter_type TEXT NOT NULL, -- 'process' | 'http' | 'claude' | 'opencode' | 'openclaw'
  adapter_config JSONB, -- blob opaco específico do adapter
  reports_to UUID FK → agents (nullable = CEO),
  capabilities TEXT, -- descrição para discovery
  status TEXT DEFAULT 'active', -- active | paused | terminated
  budget_limit NUMERIC(12,2),
  heartbeat_schedule TEXT, -- cron expression
  context_mode TEXT DEFAULT 'thin', -- 'thin' | 'fat'
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
)

-- Tasks
tasks (
  id UUID PK,
  company_id UUID FK → companies,
  parent_id UUID FK → tasks (nullable),
  agent_id UUID FK → agents (nullable = backlog),
  title TEXT NOT NULL,
  description TEXT,
  status TEXT DEFAULT 'backlog', -- backlog | available | in_progress | completed | cancelled | paused
  priority INT DEFAULT 0,
  billing_code TEXT,
  budget_spent NUMERIC(12,2) DEFAULT 0,
  completed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
)

-- Task Comments
task_comments (
  id UUID PK,
  task_id UUID FK → tasks,
  agent_id UUID FK → agents,
  content TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT now()
)

-- Heartbeats
heartbeats (
  id UUID PK,
  agent_id UUID FK → agents,
  status TEXT DEFAULT 'pending', -- pending | running | completed | failed
  mode TEXT DEFAULT 'command', -- command | fire-and-forget
  context_sent JSONB,
  result JSONB,
  started_at TIMESTAMPTZ DEFAULT now(),
  completed_at TIMESTAMPTZ,
  logs TEXT
)

-- Budgets
budgets (
  id UUID PK,
  company_id UUID FK → companies,
  agent_id UUID FK → agents (nullable = company-level),
  period_start DATE NOT NULL,
  period_end DATE NOT NULL,
  limit_tokens BIGINT,
  limit_dollars NUMERIC(12,2),
  spent_tokens BIGINT DEFAULT 0,
  spent_dollars NUMERIC(12,2) DEFAULT 0,
  alert_at NUMERIC(3,2) DEFAULT 0.80, -- % para soft alert
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
)

-- Activity Log
activity_logs (
  id UUID PK,
  company_id UUID FK → companies,
  actor_id TEXT NOT NULL, -- board | agent_id
  action TEXT NOT NULL, -- 'company.created' | 'agent.created' | 'task.completed' etc
  target_type TEXT NOT NULL, -- 'company' | 'agent' | 'task' | 'budget' | 'heartbeat'
  target_id TEXT NOT NULL,
  metadata JSONB,
  created_at TIMESTAMPTZ DEFAULT now()
)

-- Agent API Keys
agent_api_keys (
  id UUID PK,
  agent_id UUID FK → agents,
  key_hash TEXT NOT NULL, -- bcrypt
  name TEXT,
  last_used_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT now()
)

-- Approval Requests
approval_requests (
  id UUID PK,
  company_id UUID FK → companies,
  request_type TEXT NOT NULL, -- 'hire_agent' | 'strategy_breakdown'
  requested_by UUID FK → agents,
  status TEXT DEFAULT 'pending', -- pending | approved | rejected
  target_data JSONB, -- dados da ação que requer aprovação
  reviewed_by TEXT, -- board
  reviewed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT now()
)
```
