#!/bin/sh
# Render runtime configuration into the served bundle. nginx's stock entrypoint
# executes /docker-entrypoint.d/*.sh before launching the master process, so
# this runs once per container start.
set -eu

: "${NITRONODE_URL:?NITRONODE_URL must be set}"

OUT=/usr/share/nginx/html/playground/env.js
envsubst < /env.js.tmpl > "$OUT"
