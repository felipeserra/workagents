const API_BASE = '/api';

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = localStorage.getItem('token');
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  };
  if (token) headers['Authorization'] = `Bearer ${token}`;

  const res = await fetch(`${API_BASE}${path}`, { ...options, headers });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error || `HTTP ${res.status}`);
  }
  return res.json();
}

export const api = {
  auth: {
    login: (email: string, password: string) =>
      request<{ token: string; user: { id: string; name: string; email: string } }>('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      }),
    register: (name: string, email: string, password: string) =>
      request<{ id: string; name: string; email: string }>('/auth/register', {
        method: 'POST',
        body: JSON.stringify({ name, email, password }),
      }),
  },
  companies: {
    list: () => request<any[]>('/companies'),
    get: (id: string) => request<any>(`/companies/${id}`),
    create: (data: { name: string; goal?: string }) =>
      request<any>('/companies', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: { name: string; goal?: string }) =>
      request<any>(`/companies/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
    delete: (id: string) =>
      request<any>(`/companies/${id}`, { method: 'DELETE' }),
  },
  agents: {
    list: (companyId?: string) =>
      request<any[]>(`/agents${companyId ? `?company_id=${companyId}` : ''}`),
    create: (data: any) =>
      request<any>('/agents', { method: 'POST', body: JSON.stringify(data) }),
    pause: (id: string) => request<any>(`/agents/${id}/pause`, { method: 'POST' }),
    resume: (id: string) => request<any>(`/agents/${id}/resume`, { method: 'POST' }),
    delete: (id: string) => request<any>(`/agents/${id}`, { method: 'DELETE' }),
  },
  tasks: {
    list: (params?: { company_id?: string; agent_id?: string; status?: string }) => {
      const q = new URLSearchParams();
      if (params?.company_id) q.set('company_id', params.company_id);
      if (params?.agent_id) q.set('agent_id', params.agent_id);
      if (params?.status) q.set('status', params.status);
      return request<any[]>(`/tasks?${q}`);
    },
    create: (data: any) =>
      request<any>('/tasks', { method: 'POST', body: JSON.stringify(data) }),
    checkout: (id: string, agentId: string) =>
      request<any>(`/tasks/${id}/checkout?agent_id=${agentId}`, { method: 'POST' }),
    complete: (id: string) =>
      request<any>(`/tasks/${id}/complete`, { method: 'POST' }),
  },
  heartbeats: {
    trigger: (agentId: string) =>
      request<any>('/heartbeats', { method: 'POST', body: JSON.stringify({ agent_id: agentId }) }),
    list: (agentId?: string) =>
      request<any[]>(`/heartbeats${agentId ? `?agent_id=${agentId}` : ''}`),
  },
  activity: {
    list: (companyId?: string) =>
      request<any[]>(`/activity${companyId ? `?company_id=${companyId}` : ''}`),
  },
};
