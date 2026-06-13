#!/bin/sh

wait_for_database() {
    echo "Waiting for database to be ready..."
    case "$DB_TYPE" in
        "postgres"|"postgresql")
            while ! nc -z $DB_HOST $DB_PORT; do
                echo "Waiting for PostgreSQL at $DB_HOST:$DB_PORT..."
                sleep 2
            done
            ;;
        "mysql"|"mariadb")
            while ! nc -z $DB_HOST $DB_PORT; do
                echo "Waiting for MySQL at $DB_HOST:$DB_PORT..."
                sleep 2
            done
            ;;
        "sqlite")
            echo "SQLite database - no connection check needed"
            ;;
        *)
            echo "Unknown database type: $DB_TYPE"
            exit 1
            ;;
    esac
    echo "Database is ready!"
}

wait_for_database

echo "Running database migrations..."
if ./monitoring migrate up; then
    echo "Migrations completed successfully!"
else
    echo "Migration failed!"
    exit 1
fi
