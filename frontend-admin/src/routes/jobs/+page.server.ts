import { getEncoreClient } from '$lib/server/encore';
import { getSessionToken } from '$lib/server/session';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals, url }) => {
	const client = getEncoreClient(await getSessionToken(locals));
	try {
		const res = await client.jobs.ListJobRuns({
			JobName: url.searchParams.get('job_name') || '',
			Status: url.searchParams.get('status') || '',
			Limit: 200,
		});
		return {
			runs: res.runs || [],
			jobName: url.searchParams.get('job_name') || '',
			status: url.searchParams.get('status') || '',
		};
	} catch {
		return { runs: [], jobName: '', status: '' };
	}
};
