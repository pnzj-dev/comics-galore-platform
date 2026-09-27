import { McpServer } from '@modelcontextprotocol/sdk/server/mcp.js';
import { WebStandardStreamableHTTPServerTransport } from '@modelcontextprotocol/sdk/server/webStandardStreamableHttp.js';
import { z } from 'zod';

// Comics Galore MCP server (protocol layer). Business logic lives in the
// Encore backend behind key-gated `/mcp/*` endpoints. The MCP client sends an
// MCP API key as a bearer token; each tool forwards it to Encore, which
// resolves the bound user + role.

export interface Env {
	BACKEND_URL: string;
}

async function callBackend(
	backendUrl: string,
	path: string,
	body: unknown,
	auth: string,
): Promise<unknown> {
	const res = await fetch(`${backendUrl}${path}`, {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
			...(auth ? { Authorization: auth } : {}),
		},
		body: JSON.stringify(body),
	});
	const text = await res.text();
	let parsed: unknown = text;
	if (text) {
		try {
			parsed = JSON.parse(text);
		} catch {
			// non-JSON body
		}
	}
	if (!res.ok) {
		const message =
			typeof parsed === 'object' && parsed !== null && 'message' in parsed
				? String((parsed as { message: unknown }).message)
				: text;
		throw new Error(`backend error (${res.status}): ${message}`);
	}
	return parsed;
}

function textResult(data: unknown) {
	return { content: [{ type: 'text' as const, text: JSON.stringify(data ?? {}) }] };
}

function registerTools(server: McpServer, backendUrl: string, auth: string) {
	server.tool(
		'moderate_comic',
		'Approve or reject a comic pending review.',
		{ action: z.enum(['approve', 'reject']), comic_id: z.string(), reason: z.string().optional() },
		async ({ action, comic_id, reason }) =>
			textResult(
				await callBackend(backendUrl, '/mcp/moderate-comic', { action, comic_id, reason: reason ?? '' }, auth),
			),
	);

	server.tool(
		'resolve_comment_flag',
		'Resolve an open comment moderation flag.',
		{ flag_id: z.string() },
		async ({ flag_id }) => textResult(await callBackend(backendUrl, '/mcp/resolve-comment-flag', { flag_id }, auth)),
	);

	server.tool(
		'write_comment',
		'Post a comment on a comic.',
		{ comic_id: z.string(), body: z.string() },
		async ({ comic_id, body }) => textResult(await callBackend(backendUrl, '/mcp/write-comment', { comic_id, body }, auth)),
	);

	server.tool(
		'list_support_tickets',
		'List support tickets, optionally filtered by status.',
		{ status: z.string().optional() },
		async ({ status }) =>
			textResult(await callBackend(backendUrl, '/mcp/list-support-tickets', { status: status ?? '' }, auth)),
	);

	server.tool(
		'reply_support_ticket',
		'Reply to a support ticket as staff.',
		{ ticket_id: z.string(), body: z.string() },
		async ({ ticket_id, body }) =>
			textResult(await callBackend(backendUrl, '/mcp/reply-support-ticket', { ticket_id, body }, auth)),
	);

	server.tool(
		'resolve_support_ticket',
		'Mark a support ticket as resolved.',
		{ ticket_id: z.string() },
		async ({ ticket_id }) =>
			textResult(await callBackend(backendUrl, '/mcp/resolve-support-ticket', { ticket_id }, auth)),
	);

	server.tool(
		'create_comic',
		'Create a comic. Requires an uploader/admin MCP key. Uses placeholder art when cover_key/page_keys are omitted; pass publish=true to publish immediately (otherwise pending_review); pass series_title to attach to (or create) a series.',
		{
			title: z.string(),
			author: z.string().optional(),
			description: z.string().optional(),
			content_language: z.string().optional(),
			category: z.string().optional(),
			genre: z.string().optional(),
			age_rating: z.string().optional(),
			is_premium: z.boolean().optional(),
			tags: z.array(z.string()).optional(),
			reading_direction: z.string().optional(),
			cover_key: z.string().optional(),
			page_keys: z.array(z.string()).optional(),
			publish: z.boolean().optional(),
			series_title: z.string().optional(),
		},
		async (args) =>
			textResult(
				await callBackend(
					backendUrl,
					'/mcp/create-comic',
					{
						title: args.title,
						author: args.author ?? '',
						description: args.description ?? '',
						content_language: args.content_language ?? '',
						category: args.category ?? '',
						genre: args.genre ?? '',
						age_rating: args.age_rating ?? '',
						is_premium: args.is_premium ?? false,
						tags: args.tags ?? [],
						reading_direction: args.reading_direction ?? '',
						cover_key: args.cover_key ?? '',
						page_keys: args.page_keys ?? [],
						publish: args.publish ?? false,
						series_title: args.series_title ?? '',
					},
					auth,
				),
			),
	);

	server.tool(
		'list_flagged_comments',
		'List open comment moderation flags.',
		{ page: z.number().optional(), limit: z.number().optional() },
		async ({ page, limit }) =>
			textResult(
				await callBackend(backendUrl, '/mcp/list-flagged-comments', { page: page ?? 1, limit: limit ?? 20 }, auth),
			),
	);

	server.tool(
		'delete_comment',
		'Delete a comment. Requires a moderator/admin MCP key.',
		{ comment_id: z.string() },
		async ({ comment_id }) => textResult(await callBackend(backendUrl, '/mcp/delete-comment', { comment_id }, auth)),
	);

	server.tool(
		'ban_user',
		'Ban a user. Requires an admin MCP key.',
		{ user_id: z.string(), reason: z.string().optional() },
		async ({ user_id, reason }) =>
			textResult(await callBackend(backendUrl, '/mcp/ban-user', { user_id, reason: reason ?? '' }, auth)),
	);

	server.tool(
		'unban_user',
		'Unban a user. Requires an admin MCP key.',
		{ user_id: z.string() },
		async ({ user_id }) => textResult(await callBackend(backendUrl, '/mcp/unban-user', { user_id, reason: '' }, auth)),
	);

	server.tool(
		'suspend_user',
		'Suspend a user. Requires an admin MCP key.',
		{ user_id: z.string(), reason: z.string().optional() },
		async ({ user_id, reason }) =>
			textResult(await callBackend(backendUrl, '/mcp/suspend-user', { user_id, reason: reason ?? '' }, auth)),
	);
}

export default {
	async fetch(request: Request, env: Env): Promise<Response> {
		const auth = request.headers.get('Authorization') ?? '';

		const server = new McpServer({ name: 'comics-galore', version: '1.0.0' });
		registerTools(server, env.BACKEND_URL, auth);

		// Stateless transport: each request creates a fresh server + transport.
		const transport = new WebStandardStreamableHTTPServerTransport({ sessionIdGenerator: undefined });
		await server.connect(transport);
		return transport.handleRequest(request);
	},
} satisfies ExportedHandler<Env>;
