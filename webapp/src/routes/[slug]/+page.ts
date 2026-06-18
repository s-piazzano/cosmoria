import { error } from '@sveltejs/kit';
import { getProject } from '$lib/services/api';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
  const { slug } = params;

  try {
    const project = await getProject(slug);
    return { project };
  } catch (e) {
    throw error(404, 'Project not found');
  }
};
