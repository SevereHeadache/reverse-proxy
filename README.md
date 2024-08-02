# Nginx reverse proxy server

Nginx with proxy app.
Works in pair with [authservice](https://github.com/SevereHeadache/authservice)
Passes requests to proxied servers.

1. Start server  
`docker compose up -d`
2. Add configuration in conf/live
3. Create certs  
`bash init.sh example.com app.example.com`
4. Renew certs  
`bash renew.sh example.com app.example.com`
