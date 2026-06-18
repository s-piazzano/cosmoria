const BASE_URL = (import.meta.env.VITE_API_URL || 'http://localhost:8080').replace(/\/$/, '');

const getAuthHeaders = () => {
  const token = localStorage.getItem('cosmoria_token');
  if (token) {
    return { 'Authorization': `Bearer ${token}` };
  }
  return {};
};

export type AuditEntry = {
  id: string;
  action: string;
  risk_level: 'low' | 'medium' | 'high';
  timestamp: string;
};

export async function fetchAuditLogs(limit?: number) {
  const response = await fetch(`${BASE_URL}/api/audit?limit=${limit || 50}`, {
    headers: { ...getAuthHeaders() },
  });
  if (!response.ok) throw new Error('Failed to fetch audit logs');
  return await response.json();
}
