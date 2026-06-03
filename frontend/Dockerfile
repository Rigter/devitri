### STAGE 1: Build stage
FROM node:22-slim AS builder

# Install pnpm
RUN corepack enable && corepack prepare pnpm@11.5.0 --activate

# Set working directory
WORKDIR /app

# Copy dependency manifests
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./

# Install dependencies using pnpm
RUN pnpm install --frozen-lockfile

# Copy source code
COPY . .

# Public API URL baked into the static build (set via compose build.args / .env)
ARG VITE_DEVITRI_BACKEND_URL=
ENV VITE_DEVITRI_BACKEND_URL=${VITE_DEVITRI_BACKEND_URL}

# Build the SvelteKit app
RUN pnpm run build


### STAGE 2: Production stage
FROM caddy:2.7-alpine

# Copy the built static files from the builder stage
COPY --from=builder /app/build /usr/share/caddy/

# Create a Caddyfile with explicit index.html handling for SPAs.
RUN printf ":80 {\n  root * /usr/share/caddy\n  file_server\n  try_files {path} {path}/ /index.html\n  encode zstd gzip\n}\n" > /etc/caddy/Caddyfile

# Set the entrypoint to run Caddy with the configuration
CMD ["caddy", "run", "--config", "/etc/caddy/Caddyfile", "--adapter", "caddyfile"]