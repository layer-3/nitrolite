# syntax=docker/dockerfile:1.7
#
# Build context is the repository root (matches nitronode + faucet-app).
# A per-Dockerfile ignore file (playground/Dockerfile.dockerignore) opts
# `sdk/ts` back into the context since the root `.dockerignore` excludes it
# for the Go images.

FROM node:22-alpine AS builder

WORKDIR /repo

# Build the TS SDK first. playground/package.json depends on it via
# `file:../sdk/ts`, so the dist tree must exist before `npm ci` runs in
# the playground workspace.
COPY sdk/ts/package.json sdk/ts/package-lock.json ./sdk/ts/
RUN cd sdk/ts && npm ci --no-audit --no-fund
COPY sdk/ts ./sdk/ts
RUN cd sdk/ts && npm run build:ci

# Install playground deps. Done after the SDK is built so the `file:` link
# resolves to a populated dist tree.
COPY playground/package.json playground/package-lock.json ./playground/
RUN cd playground && npm ci --no-audit --no-fund

# Build the playground bundle. `npm run build` runs `tsc --noEmit && vite build`,
# emitting to playground/dist with base=/v1/playground/.
COPY playground ./playground
RUN cd playground && npm run build


FROM nginx:1.27-alpine

# envsubst (gettext) is required by the entrypoint to render env.js at start.
RUN apk add --no-cache gettext

COPY playground/deploy/nginx.conf /etc/nginx/conf.d/default.conf
COPY playground/deploy/env.js.tmpl /env.js.tmpl
COPY playground/deploy/entrypoint.sh /docker-entrypoint.d/40-playground-env.sh
RUN chmod +x /docker-entrypoint.d/40-playground-env.sh

COPY --from=builder /repo/playground/dist /usr/share/nginx/html/v1/playground

EXPOSE 8080
