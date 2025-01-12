#!/bin/bash

# Прекратить выполнение скрипта при ошибке
set -e

echo "Сборка Go-приложения..."
go build -o app .

echo "Запуск приложения..."
DB_NAME="project-sem-1"
DB_USER="validator"
DB_PASSWORD="val1dat0r"
DB_HOST="localhost"
DB_PORT="5432"

export DB_NAME
export DB_USER
export DB_PASSWORD
export DB_HOST
export DB_PORT

./app &
app_pid=$!

# Ожидаем, пока сервер запустится (проверка доступности)
until curl -s http://localhost:8080; do
    echo "Ожидание запуска сервера..."
    sleep 2
done

echo "Сервер запущен, продолжаем выполнение тестов."
