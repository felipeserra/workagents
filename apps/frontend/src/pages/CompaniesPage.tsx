import { useState, useEffect } from 'react';
import { api } from '../lib/api';
import { Plus, Trash2 } from 'lucide-react';

export function CompaniesPage() {
  const [companies, setCompanies] = useState<any[]>([]);
  const [showForm, setShowForm] = useState(false);
  const [name, setName] = useState('');
  const [goal, setGoal] = useState('');

  const load = () => api.companies.list().then(setCompanies);
  useEffect(() => { load(); }, []);

  const create = async (e: React.FormEvent) => {
    e.preventDefault();
    await api.companies.create({ name, goal });
    setName(''); setGoal(''); setShowForm(false);
    load();
  };

  const remove = async (id: string) => {
    if (confirm('Excluir empresa?')) {
      await api.companies.delete(id);
      load();
    }
  };

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <h1 style={{ fontSize: '1.5rem', fontWeight: 700 }}>Empresas</h1>
        <button className="btn btn-primary" onClick={() => setShowForm(!showForm)}>
          <Plus size={16} /> Nova Empresa
        </button>
      </div>

      {showForm && (
        <form onSubmit={create} className="card" style={{ marginBottom: 24, display: 'flex', flexDirection: 'column', gap: 12 }}>
          <div>
            <label className="label">Nome</label>
            <input className="input" value={name} onChange={e => setName(e.target.value)} placeholder="Nome da empresa" required />
          </div>
          <div>
            <label className="label">Goal</label>
            <input className="input" value={goal} onChange={e => setGoal(e.target.value)} placeholder="Objetivo principal" />
          </div>
          <div style={{ display: 'flex', gap: 8 }}>
            <button className="btn btn-primary" type="submit">Criar</button>
            <button className="btn btn-ghost" type="button" onClick={() => setShowForm(false)}>Cancelar</button>
          </div>
        </form>
      )}

      <div className="grid-3">
        {companies.map(c => (
          <div key={c.id} className="card">
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start' }}>
              <div>
                <h3 style={{ fontWeight: 600, marginBottom: 4 }}>{c.name}</h3>
                {c.goal && <p style={{ fontSize: '0.8rem', color: 'var(--muted-foreground)' }}>{c.goal}</p>}
              </div>
              <button className="btn btn-ghost" style={{ padding: 6 }} onClick={() => remove(c.id)}>
                <Trash2 size={14} />
              </button>
            </div>
            <div style={{ marginTop: 12, display: 'flex', gap: 12, fontSize: '0.8rem', color: 'var(--muted-foreground)' }}>
              <span>{c.agent_count || 0} agentes</span>
              <span>{c.task_count || 0} tarefas</span>
            </div>
          </div>
        ))}
        {companies.length === 0 && <p style={{ color: 'var(--muted-foreground)' }}>Nenhuma empresa cadastrada.</p>}
      </div>
    </div>
  );
}
