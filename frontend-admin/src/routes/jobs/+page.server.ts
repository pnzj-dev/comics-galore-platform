import { getEncoreClient } from '$lib/server/encore';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies, url }) => {
	const client = getEncoreClient(cookies.get('token'));
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
