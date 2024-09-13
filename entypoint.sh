#!/bin/sh

./migrations/migrate \
      -path ./migrations \
      -database "postgres://$POSTGRES_USERNAME:$POSTGRES_PASSWORD@$POSTGRES_HOST:$POSTGRES_PORT/$POSTGRES_DATABASE" \
      -verbose \
      up

exec "$@"
