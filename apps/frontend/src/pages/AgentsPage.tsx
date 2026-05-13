import { useState, useEffect } from 'react';
import { api } from '../lib/api';
import { Plus, PauseCircle, PlayCircle, Trash2 } from 'lucide-react';

export function AgentsPage() {
  const [agents, setAgents] = useState<any[]>([]);
  const [companies, setCompanies] = useState<any[]>([]);
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState({ company_id: '', name: '', role: '', adapter_type: 'process', adapter_config: '{}', reports_to: null as string | null });

  const load = () => Promise.all([api.agents.list(), api.companies.list()]).then(([a, c]) => { setAgents(a); setCompanies(c); });
  useEffect(() => { load(); }, []);

  const create = async (e: React.FormEvent) => {
    e.preventDefault();
    await api.agents.create(form);
    setForm({ company_id: '', name: '', role: '', adapter_type: 'process', adapter_config: '{}', reports_to: null });
    setShowForm(false);
    load();
  };

  const togglePause = async (id: string, status: string) => {
    if (status === 'active') await api.agents.pause(id);
    else await api.agents.resume(id);
    load();
  };

  const remove = async (id: string) => {
    if (confirm('Excluir agente?')) {
      await api.agents.delete(id);
      load();
    }
  };

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <h1 style={{ fontSize: '1.5rem', fontWeight: 700 }}>Agentes</h1>
        <button className="btn btn-primary" onClick={() => setShowForm(!showForm)}>
          <Plus size={16} /> Novo Agente
        </button>
      </div>

      {showForm && (
        <form onSubmit={create} className="card" style={{ marginBottom: 24, display: 'flex', flexDirection: 'column', gap: 12 }}>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            <div>
              <label className="label">Empresa</label>
              <select className="input" value={form.company_id} onChange={e => setForm({ ...form, company_id: e.target.value })} required>
                <option value="">Selecione...</option>
                {companies.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}
              </select>
            </div>
            <div>
              <label className="label">Adapter</label>
              <select className="input" value={form.adapter_type} onChange={e => setForm({ ...form, adapter_type: e.target.value })}>
                <option value="process">Process</option>
                <option value="http">HTTP</option>
              </select>
            </div>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            <div>
              <label className="label">Nome</label>
              <input className="input" value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} placeholder="Nome do agente" required />
            </div>
            <div>
              <label className="label">Role</label>
              <input className="input" value={form.role} onChange={e => setForm({ ...form, role: e.target.value })} placeholder="CEO, CTO, Dev..." required />
            </div>
          </div>
          <div>
            <label className="label">Config (JSON)</label>
            <input className="input" value={form.adapter_config} onChange={e => setForm({ ...form, adapter_config: e.target.value })} placeholder='{"command":"echo","args":["hi"]}' />
          </div>
          <div style={{ display: 'flex', gap: 8 }}>
            <button className="btn btn-primary" type="submit">Criar</button>
            <button className="btn btn-ghost" type="button" onClick={() => setShowForm(false)}>Cancelar</button>
          </div>
        </form>
      )}

      <div className="grid-3">
        {agents.map(a => (
          <div key={a.id} className="card">
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start' }}>
              <div>
                <h3 style={{ fontWeight: 600, marginBottom: 2 }}>{a.name}</h3>
                <span style={{ fontSize: '0.8rem', color: 'var(--muted-foreground)' }}>{a.role}</span>
              </div>
              <span className={`badge badge-${a.status === 'active' ? 'active' : 'paused'}`}>{a.status}</span>
            </div>
            <div style={{ marginTop: 8, fontSize: '0.8rem', color: 'var(--muted-foreground)' }}>
              <div>Adapter: {a.adapter_type}</div>
              <div>Reports to: {a.reports_to || '—'}</div>
            </div>
            <div style={{ marginTop: 12, display: 'flex', gap: 8 }}>
              <button className="btn btn-ghost" style={{ padding: '6px 12px' }} onClick={() => togglePause(a.id, a.status)}>
                {a.status === 'active' ? <PauseCircle size={14} /> : <PlayCircle size={14} />}
                {a.status === 'active' ? 'Pausar' : 'Retomar'}
              </button>
              <button className="btn btn-ghost" style={{ padding: 6 }} onClick={() => remove(a.id)}>
                <Trash2 size={14} />
              </button>
            </div>
          </div>
        ))}
        {agents.length === 0 && <p style={{ color: 'var(--muted-foreground)' }}>Nenhum agente cadastrado.</p>}
      </div>
    </div>
  );
}
