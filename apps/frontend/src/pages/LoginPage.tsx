import { useState } from 'react';
import { useAuth } from '../hooks/useAuth';

export function LoginPage() {
  const { login, register } = useAuth();
  const [isRegister, setIsRegister] = useState(false);
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      if (isRegister) {
        await register(name, email, password);
      } else {
        await login(email, password);
      }
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
      <div className="card" style={{ width: 400 }}>
        <h1 style={{ fontSize: '1.5rem', fontWeight: 700, marginBottom: 4 }}>WorkAgents</h1>
        <p style={{ color: 'var(--muted-foreground)', fontSize: '0.875rem', marginBottom: 24 }}>
          {isRegister ? 'Crie sua conta Board' : 'Acesse o Dashboard'}
        </p>

        {error && <div style={{ background: 'rgba(239,68,68,0.1)', color: '#f87171', padding: '8px 12px', borderRadius: 6, fontSize: '0.8rem', marginBottom: 16 }}>{error}</div>}

        <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          {isRegister && (
            <div>
              <label className="label">Nome</label>
              <input className="input" value={name} onChange={e => setName(e.target.value)} placeholder="Seu nome" required />
            </div>
          )}
          <div>
            <label className="label">Email</label>
            <input className="input" type="email" value={email} onChange={e => setEmail(e.target.value)} placeholder="email@exemplo.com" required />
          </div>
          <div>
            <label className="label">Senha</label>
            <input className="input" type="password" value={password} onChange={e => setPassword(e.target.value)} placeholder="••••••••" required />
          </div>
          <button className="btn btn-primary" disabled={loading} style={{ justifyContent: 'center', width: '100%' }}>
            {loading ? 'Aguarde...' : isRegister ? 'Criar Conta' : 'Entrar'}
          </button>
        </form>

        <p style={{ marginTop: 16, textAlign: 'center', fontSize: '0.8rem', color: 'var(--muted-foreground)' }}>
          {isRegister ? 'Já tem conta? ' : 'Não tem conta? '}
          <button onClick={() => { setIsRegister(!isRegister); setError(''); }}
            style={{ background: 'none', border: 'none', color: 'var(--primary)', cursor: 'pointer', fontSize: '0.8rem' }}>
            {isRegister ? 'Entrar' : 'Registrar'}
          </button>
        </p>
      </div>
    </div>
  );
}
