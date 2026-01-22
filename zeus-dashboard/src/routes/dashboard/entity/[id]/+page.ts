import type { PageLoad } from './$types';

export const load: PageLoad = ({ params, url }) => {
	return {
		entityId: params.id,
		returnUrl: url.searchParams.get('from') || '/'
	};
};
