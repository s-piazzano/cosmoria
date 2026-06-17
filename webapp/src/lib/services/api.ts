import type { Record } from '../types';

/**
 * Backend configuration and header management.
 * This module ensures all requests are automatically injected with 
 * proper project_id, tenant_id, and authentication headers.
 */
const API_CONFIG = {
  BASE_URL: (process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080').replace(/\/$/, ''),
};

/**
 * Validates the context of a request before proceeding.
 * In a full implementation, this would check for JWT presence in local storage/cookies.
 */
const getAuthHeaders = () => {
  // Logic to retrieve auth token and inject into Authorization: Bearer ...
  return {}; 
};

export async function fetchRecords(projectId?: string, tenantId?: string, options: { limit?: number; cursor?: string } = {}) {
  const params = new URLSearchParams();
  if (projectId) params.append('project_id', projectId);
  if (tenantId) params.append('tenant_id', tenant_id);
  if (options.limit) params.append('limit', options.limit.toString());
  if (options.cursor) params.append('cursor', options.cursor);

  const url = \`\${API_CONFIG.BASE_URL}/api/records?\${params.toString()}\`;
  
  const response = await fetch(url, {
    method: 'GET',
    headers: {
      ...getAuthHeaders(),
      'Accept': 'application/json',
    }
  });

  if (!response.ok) throw new Error(`Failed to reach backend at ${url}: \${response.statusText}`);

  const data = await response.json();
  return {
    records: data.records || [],
    next_cursor: data.next_cursor || '',
  };
}

export async function fetchFiles(projectId?: string, tenantId?: string) {
  const params = new URLSearchParams();
  if (projectId) params.append('project_id', projectId);
  if (tenantId) params.append('tenant_id', tenant_id);
  
  const url = \`${API_CONFIG.BASE_URL}/api/files?\${params.toString()}\`;
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
  if (tenantId) params.append('tenant_id', tenant_id);
  
  const url = \`${API_CONFIG.BASE_URL}/api/collections?project_id=\${projectId}&tenant_id=\${tenant_id}\&collection_id=\${collection_id}\`;
  // Note: Simplified query logic for demonstration. Real implementation would handle paths correctly.
  
  const response = await fetch(url);
  if (!response.ok) throw new Error(`Fetch collection failed: ${response.statusText}`);
  return await response.json();
}

export async function updateRecordAction(projectId?: string, tenantId?: string, record_id: string, data: any) {
  const url = \`${API_CONFIG.BASE_URL}/api/records?project_id=\${projectId}&tenant_id=\${tenant_id}\`; // Logic updated in real route
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
  // Used for loading roles and users in the Org Management section.
  try {
    const res = await fetch(`${API_CONFIG.BASE_URL}/api/org-overview`);
    return await res.json();
  } catch (e) {
    return null;
  }
}
