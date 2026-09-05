# frontend-public — Comics Galore web app (SvelteKit + Bun + adapter-node)
FROM oven/bun:1 AS build
WORKDIR /app

COPY package.json ./
RUN bun install

COPY . .

# Build-time (Vite) config — baked into the client bundle per environment.
ARG VITE_BACKEND_URL
ARG VITE_API_URL
ARG VITE_TURNSTILE_SITEKEY
ENV VITE_BACKEND_URL=$VITE_BACKEND_URL \
    VITE_API_URL=$VITE_API_URL \
    VITE_TURNSTILE_SITEKEY=$VITE_TURNSTILE_SITEKEY

RUN bun run build

FROM oven/bun:1
WORKDIR /app
ENV NODE_ENV=production \
    PORT=3000 \
    HOST=0.0.0.0

COPY package.json ./
RUN bun install --production --ignore-scripts

COPY --from=build /app/build ./build

CMD ["bun", "run", "start"]
