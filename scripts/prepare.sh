#!/bin/bash
# Прекратить выполнение скрипта при ошибке
set -e

echo "Установка зависимостей"
go mod tidy

echo "Подготовка базы данных..."
DB_NAME="project-sem-1"
DB_USER="validator"
DB_PASSWORD="val1dat0r"
DB_HOST="localhost"
DB_PORT="5432"

export PGPASSWORD="$DB_PASSWORD"
# Создаем базу данных
psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -d postgres -c "DROP DATABASE IF EXISTS \"$DB_NAME\";"
psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -d postgres -c "CREATE DATABASE \"$DB_NAME\";"

# Создаем таблицу
psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -d "$DB_NAME" -c "
CREATE TABLE IF NOT EXISTS prices (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL,
    price DECIMAL(10, 2) NOT NULL
);
"

echo "База данных подготовлена."
