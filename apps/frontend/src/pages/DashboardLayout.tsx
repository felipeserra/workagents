import { Outlet, NavLink, useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { LayoutDashboard, Building2, Users, ListChecks, Activity, Heart, LogOut } from 'lucide-react';

export function DashboardLayout() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div style={{ display: 'flex' }}>
      <div className="sidebar">
        <div style={{ fontSize: '1.2rem', fontWeight: 700, marginBottom: 24, padding: '0 4px' }}>
          Work<span style={{ color: 'var(--primary)' }}>Agents</span>
        </div>

        <NavLink to="/" end className={({ isActive }) => `sidebar-link ${isActive ? 'active' : ''}`}>
          <LayoutDashboard size={18} /> Dashboard
        </NavLink>
        <NavLink to="/companies" className={({ isActive }) => `sidebar-link ${isActive ? 'active' : ''}`}>
          <Building2 size={18} /> Empresas
        </NavLink>
        <NavLink to="/agents" className={({ isActive }) => `sidebar-link ${isActive ? 'active' : ''}`}>
          <Users size={18} /> Agentes
        </NavLink>
        <NavLink to="/tasks" className={({ isActive }) => `sidebar-link ${isActive ? 'active' : ''}`}>
          <ListChecks size={18} /> Tarefas
        </NavLink>
        <NavLink to="/heartbeats" className={({ isActive }) => `sidebar-link ${isActive ? 'active' : ''}`}>
          <Heart size={18} /> Heartbeats
        </NavLink>
        <NavLink to="/activity" className={({ isActive }) => `sidebar-link ${isActive ? 'active' : ''}`}>
          <Activity size={18} /> Atividades
        </NavLink>

        <div style={{ marginTop: 'auto', borderTop: '1px solid var(--border)', paddingTop: 16 }}>
          <div style={{ fontSize: '0.8rem', color: 'var(--muted-foreground)', marginBottom: 8, padding: '0 4px' }}>
            {user?.name} · {user?.email}
          </div>
          <button onClick={handleLogout} className="sidebar-link" style={{ width: '100%', border: 'none', cursor: 'pointer' }}>
            <LogOut size={18} /> Sair
          </button>
        </div>
      </div>

      <div style={{ flex: 1, padding: 32 }}>
        <Outlet />
      </div>
    </div>
  );
}
