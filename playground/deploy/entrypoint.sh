#!/bin/sh
# Render runtime configuration into the served bundle. nginx's stock entrypoint
# executes /docker-entrypoint.d/*.sh before launching the master process, so
# this runs once per container start.
set -eu

: "${NITRONODE_URL:?NITRONODE_URL must be set}"
# Default to 'true' so unset deploys keep the existing behavior.
: "${FAUCET_ENABLED:=true}"
export FAUCET_ENABLED

OUT=/usr/share/nginx/html/v1/playground/env.js
# Restrict envsubst to known variables so unrelated environment values can't
# bleed into the served config.
envsubst '${NITRONODE_URL} ${FAUCET_ENABLED}' < /env.js.tmpl > "$OUT"
