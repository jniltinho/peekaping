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

# Set default environment variables if not provided
export DB_TYPE=${DB_TYPE:-mysql}
export DB_HOST=${DB_HOST:-localhost}
export DB_PORT=${DB_PORT:-3306}
export DB_NAME=${DB_NAME:-peekaping}
export DB_USER=${DB_USER}
export DB_PASS=${DB_PASS}

# Set server configuration environment variables
export SERVER_PORT=${SERVER_PORT:-8034}
# Security: Use HTTPS by default
export CLIENT_URL=${CLIENT_URL:-http://localhost:8383}
export MODE=${MODE:-prod}
export TZ=${TZ:-UTC}

# Create .env file for the server with secure permissions
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

# Security: Set restrictive permissions on sensitive config file
chmod 600 /app/.env

# Create data directory if it doesn't exist
mkdir -p /var/lib/mysql

# Create log directory and fix permissions
mkdir -p /var/log/supervisor
chmod 755 /var/log/supervisor

# Fix ownership and permissions of MariaDB data directory (mysql user created by package)
chown -R mysql:mysql /var/lib/mysql
chmod 700 /var/lib/mysql

# Initialize MariaDB if needed
if [ ! -f /var/lib/mysql/.mysql_initialized ]; then
    echo "Initializing MariaDB..."

    # Clear data directory if it exists but is not initialized
    if [ -d /var/lib/mysql ]; then
        rm -rf /var/lib/mysql/*
    fi

    # Ensure ownership after clearing
    chown -R mysql:mysql /var/lib/mysql

    # Initialize MariaDB data directory (modern mysql_install_db)
    # --auth-root-authentication-method=normal allows root with password (we'll create app user)
    if ! mysql_install_db --user=mysql --datadir=/var/lib/mysql --auth-root-authentication-method=normal 2>/dev/null; then
        mysql_install_db --user=mysql --datadir=/var/lib/mysql
    fi

    # Start MariaDB temporarily for initialization (background/daemonize)
    echo "Starting MariaDB temporarily for initialization..."
    mysqld --datadir=/var/lib/mysql --port="$DB_PORT" --user=mysql --bind-address=0.0.0.0 --daemonize

    # Wait for MariaDB to be ready with timeout (use mysqladmin; initial root has no password after install)
    echo "Waiting for MariaDB to be ready..."
    timeout=60
    while [ $timeout -gt 0 ]; do
        if mysqladmin --user=root --host=127.0.0.1 --port="$DB_PORT" --silent ping 2>/dev/null || \
           mysqladmin --user=root --password="" --host=127.0.0.1 --port="$DB_PORT" --silent ping 2>/dev/null; then
            break
        fi
        sleep 1
        timeout=$((timeout - 1))
    done

    if [ $timeout -eq 0 ]; then
        echo "Error: MariaDB failed to start within timeout"
        exit 1
    fi

    # Create database and application user (securely via heredoc)
    echo "Creating database and user for application..."
    mysql -u root -h 127.0.0.1 -P "$DB_PORT" <<EOSQL
CREATE USER IF NOT EXISTS '$DB_USER'@'%' IDENTIFIED BY '$DB_PASS';
CREATE DATABASE IF NOT EXISTS \`$DB_NAME\`;
GRANT ALL PRIVILEGES ON \`$DB_NAME\`.* TO '$DB_USER'@'%';
FLUSH PRIVILEGES;
EOSQL

    # Stop the temporary MariaDB instance
    mysqladmin -u root -h 127.0.0.1 -P "$DB_PORT" shutdown || true

    # Mark as initialized
    touch /var/lib/mysql/.mysql_initialized
    chown mysql:mysql /var/lib/mysql/.mysql_initialized
    echo "MariaDB initialization completed!"
fi

# Start MariaDB for migrations
echo "Starting MariaDB for migrations..."
mysqld --datadir=/var/lib/mysql --port="$DB_PORT" --user=mysql --bind-address=0.0.0.0 --daemonize

# Wait for MariaDB to be ready with timeout (now using the application credentials if possible, fallback to root)
echo "Waiting for MariaDB to be ready for migrations..."
timeout=30
while [ $timeout -gt 0 ]; do
    if mysqladmin --user="$DB_USER" --password="$DB_PASS" --host=127.0.0.1 --port="$DB_PORT" --silent ping 2>/dev/null || \
       mysqladmin --user=root --host=127.0.0.1 --port="$DB_PORT" --silent ping 2>/dev/null; then
        break
    fi
    sleep 1
    timeout=$((timeout - 1))
done

if [ $timeout -eq 0 ]; then
    echo "Error: MariaDB failed to start for migrations within timeout"
    exit 1
fi

# Run database migrations
echo "Running database migrations..."
cd /app/server
if ./run-migrations.sh; then
    echo "Migrations completed successfully!"
else
    echo "ERROR: Migration failed!"
    # Attempt clean shutdown before exit
    mysqladmin -u "$DB_USER" --password="$DB_PASS" -h 127.0.0.1 -P "$DB_PORT" shutdown || true
    exit 1
fi

# Stop MariaDB after migrations (supervisor will start it again under its control)
echo "Stopping MariaDB after migrations..."
mysqladmin -u "$DB_USER" --password="$DB_PASS" -h 127.0.0.1 -P "$DB_PORT" shutdown || true

# Start supervisor
echo "Starting supervisor..."

# Note: Environment variables are passed to supervisor processes via the environment= directive
# in the supervisor configuration, so they remain available to the server even if cleared here
exec /usr/bin/supervisord -c /etc/supervisor/conf.d/supervisord.conf
