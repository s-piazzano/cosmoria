export type Record = { id: string; collection_id: string; data: any; };
export type User = { id: string; email: string; };
export type Project = { id: string; name: string; };
export type Tenant = { id: string; name: string; };
export type Collection = { id: string; name: string; schema: any; };
export type Role = { id: string; name: string; permissions: Array<{resource: string, action: string}>; };
export type AuditEntry = { id: string; action: string; resource_id?: string; risk_level: 'low' | 'medium' | 'high'; timestamp: string; };
export type FileSystem = { id: string; filename: string; size: number; mime_type: string; s3_key: string; };
