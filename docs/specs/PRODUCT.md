# WorkAgents — Product Definition

## O Que É

WorkAgents é o control plane para empresas de IA autônomas. Uma instância pode rodar múltiplas empresas.

## Core Concepts

### Company
Uma company tem:
- **Goal** — razão de existir
- **Employees** — todo employee é um agente de IA
- **Org structure** — quem reporta para quem
- **Revenue & expenses** — rastreado no nível da company
- **Task hierarchy** — todo trabalho traça de volta ao goal da company

### Employees & Agents
Cada employee tem:
- **Adapter type + config** — como o agente roda (OpenClaw, Claude Code, etc.)
- **Role & reporting** — título, para quem reporta, quem reporta para ele
- **Capabilities description** — o que o agente faz

### Agent Execution
Dois modos de heartbeat:
1. **Run a command** — Paperclip executa um processo e monitora
2. **Fire and forget** — dispara e não acompanha

### Board Governance
Toda Company tem um Board — a camada de supervisão humana.
- Approval gates: novas contratações, breakdown estratégico inicial
- Powers sempre disponíveis: modificar orçamentos, pausar/retomar agentes, override de decisões

## V1 Product Decisions

| Topic | V1 Decision |
|-------|-------------|
| Tenancy | Single-tenant, multi-company data model |
| Board | Single human operator |
| Org graph | Strict tree (reports_to nullable root) |
| Task ownership | Single assignee, atomic checkout |
| Agent adapters | process e http |
| Budget period | Billing cycle |
| Auth | Mode-dependent (local_trusted / authenticated) |
