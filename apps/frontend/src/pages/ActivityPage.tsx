import { useState, useEffect } from 'react';
import { api } from '../lib/api';

export function ActivityPage() {
  const [activity, setActivity] = useState<any[]>([]);

  useEffect(() => { api.activity.list().then(setActivity); }, []);

  return (
    <div>
      <h1 style={{ fontSize: '1.5rem', fontWeight: 700, marginBottom: 24 }}>Atividades</h1>

      <div className="card">
        {activity.map((a: any) => (
          <div key={a.id} style={{ display: 'flex', gap: 12, padding: '12px 0', borderBottom: '1px solid var(--border)' }}>
            <div style={{ width: 8, height: 8, borderRadius: '50%', background: 'var(--primary)', marginTop: 6, flexShrink: 0 }} />
            <div style={{ flex: 1 }}>
              <div style={{ fontWeight: 500, fontSize: '0.9rem' }}>{a.action}</div>
              <div style={{ color: 'var(--muted-foreground)', fontSize: '0.8rem' }}>
                {a.target_type} · {a.target_id?.substring(0, 8)}…
                {a.metadata && ` · ${a.metadata}`}
              </div>
            </div>
            <div style={{ color: 'var(--muted-foreground)', fontSize: '0.8rem', whiteSpace: 'nowrap' }}>
              {a.created_at}
            </div>
          </div>
        ))}
        {activity.length === 0 && <p style={{ color: 'var(--muted-foreground)' }}>Nenhuma atividade registrada.</p>}
      </div>
    </div>
  );
}
