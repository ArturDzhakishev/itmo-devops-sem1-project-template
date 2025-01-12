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
#psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -c "DROP DATABASE IF EXISTS $DB_NAME;"
#psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -c "CREATE DATABASE $DB_NAME;"
psql -U "validator" -h "localhost" -p "5432" -c "DROP DATABASE IF EXISTS project-sem-1;"
psql -U "validator" -h "localhost" -p "5432" -c "CREATE DATABASE project-sem-1;"


# Создаем таблицу
#psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -d "$DB_NAME" -c "
psql -U "validator" -h "localhost" -p "5432" -d "project-sem-1" -c "
CREATE TABLE IF NOT EXISTS prices (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL,
    price DECIMAL(10, 2) NOT NULL
);
"

echo "База данных подготовлена."
