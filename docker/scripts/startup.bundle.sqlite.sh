#!/bin/sh
set -e




# Set environment variables for SQLite
export DB_TYPE=sqlite
export DB_NAME=/app/data/peekaping.db

# Set server configuration environment variables
export SERVER_PORT=${SERVER_PORT:-8034}
export CLIENT_URL=${CLIENT_URL:-http://localhost:8383}
export MODE=${MODE:-prod}
export TZ=${TZ:-UTC}

# Create data directory if it doesn't exist
mkdir -p /app/data /app/server

# Create .env file for the server
cat > /app/server/.env << EOF
SERVER_PORT=$SERVER_PORT
CLIENT_URL=$CLIENT_URL
DB_TYPE=$DB_TYPE
DB_NAME=$DB_NAME
MODE=$MODE
TZ=$TZ
EOF
chmod 600 /app/server/.env

# Run database migrations
echo "Running database migrations..."
cd /app/server
if ./run-migrations.sh; then
    echo "Migrations completed successfully!"
else
    echo "Migration failed!"
    exit 1
fi

# Start supervisor
echo "Starting supervisor..."
exec /usr/bin/supervisord -c /etc/supervisor/conf.d/supervisord.conf
