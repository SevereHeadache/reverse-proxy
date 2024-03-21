# Nginx reverse proxy server

Nginx docker container.
Passes requests to proxied servers.

## Start server
`docker compose up -d`

## Create certs
`bash init-letsencrypt.sh example.com app.example.com`
