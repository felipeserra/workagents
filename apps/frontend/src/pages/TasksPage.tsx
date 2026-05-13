import { useState, useEffect } from 'react';
import { api } from '../lib/api';
import { Plus, CheckCircle, ArrowRightCircle } from 'lucide-react';

export function TasksPage() {
  const [tasks, setTasks] = useState<any[]>([]);
  const [agents, setAgents] = useState<any[]>([]);
  const [companies, setCompanies] = useState<any[]>([]);
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState({ company_id: '', title: '', description: '' });
  const [selectedAgent, setSelectedAgent] = useState('');

  const load = () => Promise.all([api.tasks.list(), api.agents.list(), api.companies.list()]).then(([t, a, c]) => { setTasks(t); setAgents(a); setCompanies(c); });
  useEffect(() => { load(); }, []);

  const create = async (e: React.FormEvent) => {
    e.preventDefault();
    await api.tasks.create({ ...form, agent_id: selectedAgent || null });
    setForm({ company_id: '', title: '', description: '' });
    setSelectedAgent('');
    setShowForm(false);
    load();
  };

  const checkout = async (taskId: string) => {
    const agentId = prompt('Agent ID para checkout:');
    if (!agentId) return;
    await api.tasks.checkout(taskId, agentId);
    load();
  };

  const complete = async (taskId: string) => {
    await api.tasks.complete(taskId);
    load();
  };

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <h1 style={{ fontSize: '1.5rem', fontWeight: 700 }}>Tarefas</h1>
        <button className="btn btn-primary" onClick={() => setShowForm(!showForm)}>
          <Plus size={16} /> Nova Tarefa
        </button>
      </div>

      {showForm && (
        <form onSubmit={create} className="card" style={{ marginBottom: 24, display: 'flex', flexDirection: 'column', gap: 12 }}>
          <div>
            <label className="label">Empresa</label>
            <select className="input" value={form.company_id} onChange={e => setForm({ ...form, company_id: e.target.value })} required>
              <option value="">Selecione...</option>
              {companies.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}
            </select>
          </div>
          <div>
            <label className="label">Título</label>
            <input className="input" value={form.title} onChange={e => setForm({ ...form, title: e.target.value })} placeholder="Título da tarefa" required />
          </div>
          <div>
            <label className="label">Descrição</label>
            <textarea className="input" value={form.description} onChange={e => setForm({ ...form, description: e.target.value })} placeholder="Descrição..." rows={3} />
          </div>
          <div>
            <label className="label">Agente (opcional)</label>
            <select className="input" value={selectedAgent} onChange={e => setSelectedAgent(e.target.value)}>
              <option value="">Backlog (sem assignee)</option>
              {agents.filter(a => a.status === 'active').map(a => <option key={a.id} value={a.id}>{a.name} ({a.role})</option>)}
            </select>
          </div>
          <div style={{ display: 'flex', gap: 8 }}>
            <button className="btn btn-primary" type="submit">Criar</button>
            <button className="btn btn-ghost" type="button" onClick={() => setShowForm(false)}>Cancelar</button>
          </div>
        </form>
      )}

      <div className="grid-3">
        {tasks.map(t => (
          <div key={t.id} className="card">
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start' }}>
              <div>
                <h3 style={{ fontWeight: 600, marginBottom: 2, fontSize: '0.95rem' }}>{t.title}</h3>
                {t.description && <p style={{ fontSize: '0.8rem', color: 'var(--muted-foreground)', marginTop: 4 }}>{t.description}</p>}
              </div>
              <span className={`badge badge-${t.status === 'completed' ? 'completed' : t.status === 'in_progress' ? 'active' : 'pending'}`}>{t.status}</span>
            </div>
            <div style={{ marginTop: 8, fontSize: '0.8rem', color: 'var(--muted-foreground)' }}>
              Prioridade: {t.priority || 0}
            </div>
            <div style={{ marginTop: 12, display: 'flex', gap: 8 }}>
              {t.status === 'backlog' && (
                <button className="btn btn-ghost" style={{ padding: '6px 12px', fontSize: '0.8rem' }} onClick={() => checkout(t.id)}>
                  <ArrowRightCircle size={14} /> Checkout
                </button>
              )}
              {t.status === 'in_progress' && (
                <button className="btn btn-primary" style={{ padding: '6px 12px', fontSize: '0.8rem' }} onClick={() => complete(t.id)}>
                  <CheckCircle size={14} /> Completar
                </button>
              )}
            </div>
          </div>
        ))}
        {tasks.length === 0 && <p style={{ color: 'var(--muted-foreground)' }}>Nenhuma tarefa.</p>}
      </div>
    </div>
  );
}
