# Проект: сервис для управления задачами

Этот проект содержит многоконтейнерное приложение, использующее PostgreSQL, Kafka и Zookeeper, развертываемое через Docker Compose.

## Сервисы

- **app** - основное приложение (порт 3000)
- **db** - база данных PostgreSQL 15
- **zookeeper** - координационный сервис для Kafka
- **kafka** - брокер сообщений Apache Kafka

## Запуск проекта

1. Убедитесь, что установлены Docker и Docker Compose
2. Выполните команду:
```bash
make docker_up
```
Приложение будет доступно по адресу: http://localhost:3000

Работа с базой данных
Подключение к контейнеру базы данных
Чтобы подключиться к контейнеру PostgreSQL для выполнения SQL-запросов:


просмотрите запущенные контейнеры используя
```bash
docker ps
```

найдите id контейнера db-1 и выполните
```bash
docker exec -it <container_id> bash
```

зайдите в postgreSQL
```bash
psql -U postgres -d betera
```

и выполните
```sql
CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(100) NOT NULL DEFAULT 'created'
);
```



