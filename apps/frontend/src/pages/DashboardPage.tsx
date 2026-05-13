import { useState, useEffect } from 'react';
import { api } from '../lib/api';
import { Building2, Users, ListChecks, Activity } from 'lucide-react';

export function DashboardPage() {
  const [companies, setCompanies] = useState<any[]>([]);
  const [agents, setAgents] = useState<any[]>([]);
  const [tasks, setTasks] = useState<any[]>([]);
  const [activity, setActivity] = useState<any[]>([]);

  useEffect(() => {
    api.companies.list().then(setCompanies).catch(() => {});
    api.agents.list().then(setAgents).catch(() => {});
    api.tasks.list().then(setTasks).catch(() => {});
    api.activity.list().then(setActivity).catch(() => {});
  }, []);

  const stats = [
    { icon: <Building2 size={20} />, label: 'Empresas', value: companies.length, color: '#60A5FA' },
    { icon: <Users size={20} />, label: 'Agentes', value: agents.length, color: '#A78BFA' },
    { icon: <ListChecks size={20} />, label: 'Tarefas', value: tasks.length, color: '#34D399' },
    { icon: <Activity size={20} />, label: 'Atividades', value: activity.length, color: '#FBBF24' },
  ];

  return (
    <div>
      <h1 style={{ fontSize: '1.5rem', fontWeight: 700, marginBottom: 24 }}>Dashboard</h1>

      <div className="grid-3" style={{ marginBottom: 32 }}>
        {stats.map((s, i) => (
          <div key={i} className="card" style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
            <div style={{ width: 44, height: 44, borderRadius: 10, background: `${s.color}15`, display: 'flex', alignItems: 'center', justifyContent: 'center', color: s.color }}>
              {s.icon}
            </div>
            <div>
              <div style={{ fontSize: '1.5rem', fontWeight: 700 }}>{s.value}</div>
              <div style={{ fontSize: '0.8rem', color: 'var(--muted-foreground)' }}>{s.label}</div>
            </div>
          </div>
        ))}
      </div>

      <div className="grid-2">
        <div className="card">
          <h2 style={{ fontSize: '1rem', fontWeight: 600, marginBottom: 16 }}>Últimas Atividades</h2>
          {activity.slice(0, 10).map((a: any) => (
            <div key={a.id} style={{ display: 'flex', justifyContent: 'space-between', padding: '8px 0', borderBottom: '1px solid var(--border)', fontSize: '0.8rem' }}>
              <span style={{ color: 'var(--primary)' }}>{a.action}</span>
              <span style={{ color: 'var(--muted-foreground)' }}>{a.target_type}</span>
            </div>
          ))}
          {activity.length === 0 && <p style={{ color: 'var(--muted-foreground)', fontSize: '0.875rem' }}>Nenhuma atividade ainda.</p>}
        </div>

        <div className="card">
          <h2 style={{ fontSize: '1rem', fontWeight: 600, marginBottom: 16 }}>Tarefas Recentes</h2>
          {tasks.slice(0, 10).map((t: any) => (
            <div key={t.id} style={{ display: 'flex', justifyContent: 'space-between', padding: '8px 0', borderBottom: '1px solid var(--border)', fontSize: '0.8rem' }}>
              <span>{t.title}</span>
              <span className={`badge badge-${t.status === 'completed' ? 'completed' : t.status === 'in_progress' ? 'active' : 'pending'}`}>{t.status}</span>
            </div>
          ))}
          {tasks.length === 0 && <p style={{ color: 'var(--muted-foreground)', fontSize: '0.875rem' }}>Nenhuma tarefa ainda.</p>}
        </div>
      </div>
    </div>
  );
}
