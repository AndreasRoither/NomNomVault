# ---- base workspace setup
FROM node:22-alpine

WORKDIR /workspace

RUN corepack enable

# Copy workspace manifests so dependency install can be cached
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml .npmrc ./
COPY frontend/apps/grocery-web/package.json frontend/apps/grocery-web/package.json
COPY frontend/packages/api-client/package.json frontend/packages/api-client/package.json
COPY frontend/packages/ui/package.json frontend/packages/ui/package.json
COPY frontend/packages/app-shell/package.json frontend/packages/app-shell/package.json

RUN pnpm install --frozen-lockfile

# ---- application source
COPY frontend ./frontend

WORKDIR /workspace/frontend/apps/grocery-web

# ---- production build
ENV NODE_ENV=production
ENV HOST=0.0.0.0
ENV PORT=3001

RUN pnpm build

# ---- runtime
EXPOSE 3001

CMD ["node", ".output/server/index.mjs"]
