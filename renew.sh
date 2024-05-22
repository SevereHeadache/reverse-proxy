#!/bin/bash

domains=("$@")
dry_run=0 # Dry run

if [ ${#domains[@]} -eq 0 ]; then
  echo "Error: No domains"
  exit 1
fi

for domain in "${domains[@]}"; do
  echo "### Renewing Let's Encrypt certificate for $domain ..."
  # Enable dry run mode if needed
  if [ $dry_run != "0" ]; then dry_arg="--dry-run"; fi
  docker compose run --rm --entrypoint "\
    certbot renew \
      --cert-name $domain \
      $dry_arg" certbot
  echo
done

echo "### Restarting nginx ..."
docker compose up --force-recreate -d nginx
