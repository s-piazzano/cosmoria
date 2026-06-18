import { writable } from 'svelte/store';

export const user = writable<any>(null);
export const projectInfo = writable<{id: string, name: string} | null>(null);
export const tenantInfo = writable<{id: string, name: string} | null>(null);
export const currentSelection = writable<string>('dashboard');
export const activeTab = writable<string>('dashboard');

export async function setSessionData(userData: any, projectId: string, tenantId: string) {
    user.set(userData);
    projectInfo.set({id: projectId, name: ''});
    tenantInfo.set({id: tenantId, name: ''});
}
