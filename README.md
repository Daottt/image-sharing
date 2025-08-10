# Image Sharing

API для обмена изображениями, написанное на GO с использованием chi для роутеров и sqlc для работы с базой данных.
Для аутентификации использует https://github.com/Daottt/sso

## Зависимости

- GO 1.24
- PostgreSQL 16

## Установка и запуск

1. Клонируем репозиторий
    ```bash
    git clone https://github.com/Daottt/image-sharing.git
    cd image-sharing
    ```
2. Запускаем контейнеры
    ```bash
    docker compose up
    ```
