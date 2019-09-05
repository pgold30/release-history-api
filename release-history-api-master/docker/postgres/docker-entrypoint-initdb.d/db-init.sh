#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE ROLE releasehistory WITH ENCRYPTED PASSWORD 'releasehistorylocal';
    CREATE DATABASE releasehistory;
EOSQL

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" releasehistory <<-EOSQL
    GRANT ALL ON ALL TABLES IN SCHEMA public TO releasehistory;
    ALTER ROLE releasehistory WITH LOGIN;
EOSQL
