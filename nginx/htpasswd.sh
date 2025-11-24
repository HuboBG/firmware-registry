#!/bin/sh
set -e

if [ -n "$NGINX_BASIC_AUTH_USER" ] && [ -n "$NGINX_BASIC_AUTH_PASS" ]; then
  echo "Creating htpasswd for UI basic auth..."
  apk add --no-cache apache2-utils >/dev/null 2>&1
  htpasswd -bc /etc/nginx/.htpasswd "$NGINX_BASIC_AUTH_USER" "$NGINX_BASIC_AUTH_PASS"
else
  echo "No basic auth env vars set; leaving UI unprotected by basic auth."
fi
