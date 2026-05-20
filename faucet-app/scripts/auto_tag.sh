#!/bin/sh

last_rc_tag=$(git ls-remote -t | grep -Eo 'v[0-9]+\.[0-9]+\.[0-9]+-rc\.[0-9]+$' | sort -V -r | head -n 1)
last_stable_tag=$(git ls-remote -t | grep -Eo 'v[0-9]+\.[0-9]+\.[0-9]+$' | sort -V -r | head -n 1)

new_rc_tag=""
if [ -z "$last_rc_tag" ] && [ -z "$last_stable_tag" ]; then
  # No tags found
  new_rc_tag="v0.0.0-rc.0"
elif [ -z "$last_stable_tag" ]; then
  # No stable tags found
  new_rc_tag="${last_rc_tag%.*}.$((${last_rc_tag##*.}+1))"
elif [ "${last_rc_tag%-*}" = "${last_stable_tag}" ]; then
  # Last RC tag is the latest stable tag
  new_rc_tag="${last_stable_tag}-rc.$((${last_rc_tag##*.}+1))"
else
  # Last RC tag is not the latest stable tag
  new_rc_tag="${last_stable_tag}-rc.0"
fi

git tag "$new_rc_tag"
git push origin "$new_rc_tag"
