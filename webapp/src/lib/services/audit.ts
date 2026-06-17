type auditEntry = {
  id: string;
  action: string;
  risk_level: 'low' | 'medium' | 'high';
  timestamp: string;
};

export async function fetchAuditLogs(limit?: number) {
  try {
    const response = await fetch(`http://localhost:8080/api/audit?limit=${limit || 50}`);
    if (!response.ok) throw new Error('Failed to fetch audit logs');
    return await response.json();
  } catch (e) {
    console.error(e);
    return [];
  }
}
