# Comics Galore MCP server (Cloudflare Worker)

Remote MCP server (Streamable HTTP) that fronts the Comics Galore Encore
backend's key-gated `/mcp/*` endpoints. Business logic + role checks live in
Encore; this Worker is the MCP protocol layer.

## Architecture

```
MCP client (Claude/Cursor/etc.)
   │  Authorization: Bearer <mcp-key>
   ▼
Cloudflare Worker (createMcpHandler)  ── https://<worker>.workers.dev/mcp
   │  forwards the bearer key to the Encore backend
   ▼
Encore /mcp/* (public, key-gated)  →  resolves mcp_keys → user + role
   │
   ▼
comics / social internal endpoints (moderate, comment, support)
```

## Tools

| Tool | Encore endpoint | Role |
|------|-----------------|------|
| `moderate_comic` | `POST /mcp/moderate-comic` | moderator/admin |
| `resolve_comment_flag` | `POST /mcp/resolve-comment-flag` | moderator/admin |
| `write_comment` | `POST /mcp/write-comment` | any |
| `list_support_tickets` | `POST /mcp/list-support-tickets` | moderator/admin |
| `reply_support_ticket` | `POST /mcp/reply-support-ticket` | moderator/admin |
| `resolve_support_ticket` | `POST /mcp/resolve-support-ticket` | moderator/admin |

## Issuing an MCP key

Keys are stored (hashed) in the `mcpdb.mcp_keys` table, bound to an internal
user. Create and manage them from the **admin panel** (`Admin → MCP Keys`):
the full key is shown **once** on creation; afterwards only the last 4
characters are displayed.

## Deploy

```bash
cd mcp-worker
bun install
# set the backend base URL (per environment)
npx wrangler secret put BACKEND_URL
npx wrangler deploy
```

## Connect

MCP clients connect to `https://<worker>.<account>.workers.dev/mcp` and send
the MCP key as a bearer token. Example (`claude` / generic MCP client):

```json
{
  "mcpServers": {
    "comics-galore": {
      "type": "http",
      "url": "https://<worker>.<account>.workers.dev/mcp",
      "headers": { "Authorization": "Bearer <mcp-key>" }
    }
  }
}
```

Optional: enable Logto OAuth for the client→Worker hop (see Cloudflare
`workers-oauth-provider` + "Secure MCP servers"). The bearer-key path above
works for all clients today.
