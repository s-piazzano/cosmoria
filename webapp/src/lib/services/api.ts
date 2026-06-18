import type { ProjectWithRole, AdminUser, OverviewData, Tenant } from '../types';

const API_CONFIG = {
  BASE_URL: (import.meta.env.VITE_API_URL || '').replace(/\/$/, ''),
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

export async function validateToken(): Promise<AdminUser | null> {
  try {
    const response = await fetch(`${API_CONFIG.BASE_URL}/api/admin/me`, {
      headers: { ...getAuthHeaders(), 'Accept': 'application/json' },
    });
    if (!response.ok) return null;
    return await response.json();
  } catch {
    return null;
  }
}

export async function fetchProjects(): Promise<ProjectWithRole[]> {
  const response = await fetch(`${API_CONFIG.BASE_URL}/api/admin/projects`, {
    headers: { ...getAuthHeaders(), 'Accept': 'application/json' },
  });
  if (!response.ok) throw new Error('Failed to fetch projects');
  return await response.json();
}

export async function createProject(name: string, multitenancyEnabled: boolean = false): Promise<ProjectWithRole> {
  const response = await fetch(`${API_CONFIG.BASE_URL}/api/admin/projects`, {
    method: 'POST',
    headers: { ...getAuthHeaders(), 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, multitenancy_enabled: multitenancyEnabled }),
  });
  if (!response.ok) throw new Error('Failed to create project');
  return await response.json();
}

export async function updateProject(slug: string, patch: { name?: string; multitenancy_enabled?: boolean }): Promise<ProjectWithRole> {
  const response = await fetch(`${API_CONFIG.BASE_URL}/api/admin/projects/${encodeURIComponent(slug)}`, {
    method: 'PUT',
    headers: { ...getAuthHeaders(), 'Content-Type': 'application/json' },
    body: JSON.stringify(patch),
  });
  if (!response.ok) throw new Error('Failed to update project');
  return await response.json();
}

export async function getProject(slug: string): Promise<ProjectWithRole> {
  const response = await fetch(`${API_CONFIG.BASE_URL}/api/admin/projects/${encodeURIComponent(slug)}`, {
    headers: { ...getAuthHeaders(), 'Accept': 'application/json' },
  });
  if (!response.ok) throw new Error('Project not found');
  return await response.json();
}

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

export async function listTenants(slug: string): Promise<Tenant[]> {
  const response = await fetch(`${API_CONFIG.BASE_URL}/api/admin/projects/${encodeURIComponent(slug)}/tenants`, {
    headers: { ...getAuthHeaders(), 'Accept': 'application/json' },
  });
  if (!response.ok) throw new Error('Failed to fetch tenants');
  return await response.json();
}

export async function createTenant(slug: string, name: string): Promise<Tenant> {
  const response = await fetch(`${API_CONFIG.BASE_URL}/api/admin/projects/${encodeURIComponent(slug)}/tenants`, {
    method: 'POST',
    headers: { ...getAuthHeaders(), 'Content-Type': 'application/json' },
    body: JSON.stringify({ name }),
  });
  if (!response.ok) throw new Error('Failed to create tenant');
  return await response.json();
}

export async function deleteTenant(slug: string, tenantId: string): Promise<void> {
  const response = await fetch(`${API_CONFIG.BASE_URL}/api/admin/projects/${encodeURIComponent(slug)}/tenants/${encodeURIComponent(tenantId)}`, {
    method: 'DELETE',
    headers: { ...getAuthHeaders() },
  });
  if (!response.ok) throw new Error('Failed to delete tenant');
}

export async function fetchOverview(slug: string): Promise<OverviewData> {
  const response = await fetch(`${API_CONFIG.BASE_URL}/api/admin/projects/${encodeURIComponent(slug)}/overview`, {
    headers: { ...getAuthHeaders(), 'Accept': 'application/json' },
  });
  if (!response.ok) throw new Error('Failed to fetch project overview');
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
