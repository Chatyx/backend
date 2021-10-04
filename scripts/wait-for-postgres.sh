#!/bin/bash

until docker exec scht-postgres_test pg_isready -q; do
  >&2 echo "Postgres is unavailable - sleeping..."
  sleep 1
done

>&2 echo "Postgres is up!"