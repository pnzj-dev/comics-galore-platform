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
