#!/bin/sh
set -e

validate_env_vars() {
    local errors=0

    if [ -z "$DB_USER" ]; then
        echo "ERROR: DB_USER is required and must be set"
        errors=1
    fi

    if [ -z "$DB_PASS" ]; then
        echo "ERROR: DB_PASS is required and must be set"
        errors=1
    fi

    if [ $errors -eq 1 ]; then
        echo "Environment validation failed. Please fix the above errors."
        exit 1
    fi

    echo "Environment validation passed."
}

validate_env_vars

export DB_TYPE=${DB_TYPE:-mysql}
export DB_HOST=${DB_HOST:-localhost}
export DB_PORT=${DB_PORT:-3306}
export DB_NAME=${DB_NAME:-peekaping}
export SERVER_PORT=${SERVER_PORT:-8034}
export CLIENT_URL=${CLIENT_URL:-http://localhost:8383}
export MODE=${MODE:-prod}
export TZ=${TZ:-UTC}

SOCKET=/run/mysqld/mysqld.sock

# Create required directories
mkdir -p /var/lib/mysql /var/log/supervisor /run/mysqld
chown -R mysql:mysql /var/lib/mysql /run/mysqld
chmod 700 /var/lib/mysql

# Create .env file for the server
cat > /app/.env << EOF
SERVER_PORT=$SERVER_PORT
CLIENT_URL=$CLIENT_URL
DB_TYPE=$DB_TYPE
DB_HOST=$DB_HOST
DB_PORT=$DB_PORT
DB_NAME=$DB_NAME
DB_USER=$DB_USER
DB_PASS=$DB_PASS
MODE=$MODE
TZ=$TZ
EOF
chmod 600 /app/.env

# Root connects via unix socket (no password needed with MariaDB unix_socket auth).
mariadb_start() {
    mysqld --datadir=/var/lib/mysql --port="$DB_PORT" --user=mysql --bind-address=0.0.0.0 &
    echo "Waiting for MariaDB socket..."
    local timeout=60
    while [ $timeout -gt 0 ]; do
        if mysql --user=root --socket="$SOCKET" -e "SELECT 1" >/dev/null 2>&1; then
            echo "MariaDB is ready."
            return 0
        fi
        sleep 1
        timeout=$((timeout - 1))
    done
    echo "ERROR: MariaDB failed to start within timeout"
    exit 1
}

mariadb_stop() {
    mysqladmin --user=root --socket="$SOCKET" shutdown 2>/dev/null || true
    local timeout=15
    while [ $timeout -gt 0 ]; do
        if ! mysql --user=root --socket="$SOCKET" -e "SELECT 1" >/dev/null 2>&1; then
            return 0
        fi
        sleep 1
        timeout=$((timeout - 1))
    done
}

# Initialize MariaDB data directory on first run
if [ ! -f /var/lib/mysql/.mysql_initialized ]; then
    echo "Initializing MariaDB data directory..."
    rm -rf /var/lib/mysql/*
    chown -R mysql:mysql /var/lib/mysql
    mysql_install_db --user=mysql --datadir=/var/lib/mysql
    touch /var/lib/mysql/.mysql_initialized
    chown mysql:mysql /var/lib/mysql/.mysql_initialized
    echo "MariaDB data directory initialized."
fi

# Start MariaDB and ensure database/user exist (idempotent — safe on every start)
mariadb_start

echo "Ensuring database and user exist..."
mysql --user=root --socket="$SOCKET" <<EOSQL
CREATE USER IF NOT EXISTS '$DB_USER'@'%'         IDENTIFIED BY '$DB_PASS';
CREATE USER IF NOT EXISTS '$DB_USER'@'localhost' IDENTIFIED BY '$DB_PASS';
CREATE DATABASE IF NOT EXISTS \`$DB_NAME\`;
GRANT ALL PRIVILEGES ON \`$DB_NAME\`.* TO '$DB_USER'@'%';
GRANT ALL PRIVILEGES ON \`$DB_NAME\`.* TO '$DB_USER'@'localhost';
FLUSH PRIVILEGES;
EOSQL

# Run migrations
echo "Running database migrations..."
cd /app/server
if ./run-migrations.sh; then
    echo "Migrations completed successfully!"
else
    echo "ERROR: Migration failed!"
    mariadb_stop
    exit 1
fi

mariadb_stop

# Hand off to supervisord (manages MariaDB, Redis, api, engine)
echo "Starting supervisor..."
exec /usr/bin/supervisord -c /etc/supervisor/conf.d/supervisord.conf
