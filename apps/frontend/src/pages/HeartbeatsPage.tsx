import { useState, useEffect } from 'react';
import { api } from '../lib/api';
import { Play } from 'lucide-react';

export function HeartbeatsPage() {
  const [agents, setAgents] = useState<any[]>([]);
  const [heartbeats, setHeartbeats] = useState<any[]>([]);
  const [selectedAgent, setSelectedAgent] = useState('');

  useEffect(() => { api.agents.list().then(setAgents); }, []);

  const loadHeartbeats = (agentId: string) => {
    setSelectedAgent(agentId);
    api.heartbeats.list(agentId).then(setHeartbeats);
  };

  const trigger = async (agentId: string) => {
    await api.heartbeats.trigger(agentId);
    loadHeartbeats(agentId);
  };

  return (
    <div>
      <h1 style={{ fontSize: '1.5rem', fontWeight: 700, marginBottom: 24 }}>Heartbeats</h1>

      <div className="grid-3" style={{ marginBottom: 24 }}>
        {agents.filter(a => a.status === 'active').map(a => (
          <div key={a.id} className="card" style={{ cursor: 'pointer', border: selectedAgent === a.id ? '1px solid var(--primary)' : '' }}
            onClick={() => loadHeartbeats(a.id)}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <div>
                <h3 style={{ fontWeight: 600, fontSize: '0.95rem' }}>{a.name}</h3>
                <span style={{ fontSize: '0.8rem', color: 'var(--muted-foreground)' }}>{a.role}</span>
              </div>
              <button className="btn btn-primary" style={{ padding: '6px 12px' }} onClick={e => { e.stopPropagation(); trigger(a.id); }}>
                <Play size={14} />
              </button>
            </div>
          </div>
        ))}
      </div>

      {selectedAgent && (
        <div className="card">
          <h2 style={{ fontSize: '1rem', fontWeight: 600, marginBottom: 16 }}>Histórico</h2>
          {heartbeats.map((hb: any) => (
            <div key={hb.id} style={{ display: 'flex', justifyContent: 'space-between', padding: '10px 0', borderBottom: '1px solid var(--border)', fontSize: '0.85rem' }}>
              <div>
                <span className={`badge badge-${hb.status === 'completed' ? 'completed' : hb.status === 'running' ? 'active' : 'pending'}`}
                  style={{ marginRight: 8 }}>{hb.status}</span>
                <span style={{ color: 'var(--muted-foreground)' }}>{hb.mode}</span>
              </div>
              <div style={{ color: 'var(--muted-foreground)', fontSize: '0.8rem' }}>
                {hb.started_at?.substring(11, 19)}
                {hb.completed_at && ` → ${hb.completed_at.substring(11, 19)}`}
              </div>
            </div>
          ))}
          {heartbeats.length === 0 && <p style={{ color: 'var(--muted-foreground)' }}>Nenhum heartbeat para este agente.</p>}
        </div>
      )}
    </div>
  );
}
