import type { Record } from '../types';

const API_CONFIG = {
  BASE_URL: (import.meta.env.VITE_API_URL || 'http://localhost:8080').replace(/\/$/, ''),
};

const getAuthHeaders = () => {
  const token = localStorage.getItem('cosmoria_token');
  if (token) {
    return { 'Authorization': `Bearer ${token}` };
  }
  return {};
};

export const setAuthToken = (token: string) => {
  localStorage.setItem('cosmoria_token', token);
};

export const clearAuthToken = () => {
  localStorage.removeItem('cosmoria_token');
};

export async function fetchRecords(projectId?: string, tenantId?: string, options: { limit?: number; cursor?: string } = {}) {
  const params = new URLSearchParams();
  if (projectId) params.append('project_id', projectId);
  if (tenantId) params.append('tenant_id', tenantId);
  if (options.limit) params.append('limit', options.limit.toString());
  if (options.cursor) params.append('cursor', options.cursor);

  const url = `${API_CONFIG.BASE_URL}/api/records?${params.toString()}`;

  const response = await fetch(url, {
    method: 'GET',
    headers: {
      ...getAuthHeaders(),
      'Accept': 'application/json',
    }
  });

  if (!response.ok) throw new Error(`Failed to reach backend at ${url}: ${response.statusText}`);

  const data = await response.json();
  return {
    records: data.records || [],
    next_cursor: data.next_cursor || '',
  };
}

export async function fetchFiles(projectId?: string, tenantId?: string) {
  const params = new URLSearchParams();
  if (projectId) params.append('project_id', projectId);
  if (tenantId) params.append('tenant_id', tenantId);

  const url = `${API_CONFIG.BASE_URL}/api/files?${params.toString()}`;
  const response = await fetch(url, {
    method: 'GET',
    headers: { ...getAuthHeaders(), "Accept": "application/json" }
  });

  if (!response.ok) throw new Error(`Failed to reach file service at ${url}`);
  const data = await response.json();
  return data.files || [];
}

export async function fetchCollection(projectId?: string, tenantId?: string, collection_id: string) {
  const params = new URLSearchParams();
  if (projectId) params.append('project_id', projectId);
  if (tenantId) params.append('tenant_id', tenantId);
  if (collection_id) params.append('collection_id', collection_id);

  const url = `${API_CONFIG.BASE_URL}/api/collections?${params.toString()}`;

  const response = await fetch(url, {
    headers: { ...getAuthHeaders() },
  });
  if (!response.ok) throw new Error(`Fetch collection failed: ${response.statusText}`);
  return await response.json();
}

export async function updateRecordAction(projectId?: string, tenantId?: string, record_id: string, data: any) {
  const params = new URLSearchParams();
  if (projectId) params.append('project_id', projectId);
  if (tenantId) params.append('tenant_id', tenantId);

  const url = `${API_CONFIG.BASE_URL}/api/records/${encodeURIComponent(record_id)}?${params.toString()}`;
  const response = await fetch(url, {
    method: 'PUT',
    headers: {
      ...getAuthHeaders(),
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(data)
  });
  return await response.json();
}

export async function fetchOrganizationData() {
  try {
    const res = await fetch(`${API_CONFIG.BASE_URL}/api/org-overview`, {
      headers: { ...getAuthHeaders() },
    });
    return await res.json();
  } catch (e) {
    return null;
  }
}
