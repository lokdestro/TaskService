# Проект: Docker Compose для приложения с базой данных и Kafka

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
docker-compose up -d
```
Приложение будет доступно по адресу: http://localhost:3000

Работа с базой данных
Подключение к контейнеру базы данных
Чтобы подключиться к контейнеру PostgreSQL для выполнения SQL-запросов:


```bash
docker-compose exec db psql -U user -d betera
```

и выполните
```sql
CREATE TABLE tasks (
  id SERIAL PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  description TEXT,
  status VARCHAR(100) NOT NULL
);```



