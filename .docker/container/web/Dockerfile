# syntax=docker/dockerfile:1

FROM oven/bun:1.3-alpine AS builder

WORKDIR /app

COPY package.json bun.lock ./

RUN bun install --frozen-lockfile

COPY . .

ARG VITE_API_BASE_URL

ENV VITE_API_BASE_URL=$VITE_API_BASE_URL

RUN bun run build

FROM oven/bun:1.3-alpine

WORKDIR /app

ENV NODE_ENV=production

COPY package.json bun.lock ./

RUN bun install --frozen-lockfile --production

COPY --from=builder /app/dist ./dist
COPY server.ts index.html ./

EXPOSE 3000

CMD ["bun", "server.ts"]
