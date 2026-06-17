import { writable } from 'svelte/store';

// Core application states
export const user = writable<any>(null); // Current user info
export const projectInfo = writable<{id: string, name: string} | null>(null);
export const tenantInfo = writable<{id: string, name: string} | null>(null);
export const currentSelection = writable<string>('dashboard');

// UI State for navigation and selection
export const activeTab = writable(currentSelection);

// Helper to update multiple state items at once
export async function setSessionData(userData: any, projectId: string, tenantId: string) {
    user.set(userData);
    projectInfo.set({id: projectId, name: ''}); // Name filled from API later
    tenantInfo.set({id: tenantId, name: ''});
}
