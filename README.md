## Конфигурация проекта

Для демонстрации работы используются тестовые конфигурационные файлы:
`deploy/.env.dev`
`test_config.yaml`

## Запуск проекта

### Варианты запуска

С помощью команды 

```shell
task docker-buildup 
```

#### Docker

```shell
cd deploy
```
Включить сборку:
```shell
docker compose --env-file .env.dev up --build -d
```
Без сборки:
```shell
docker compose --env-file .env.dev up -d
```
#### Локально(без поднятия БД)

```shell
go run cmd/main.go
```

## Taskfile

Помимо запуска Taskfile предоставляет различный функционал. К примеру запуск тестов:

```shell
task test
```

Также можно получить процент покрытия тестами:


```shell
task coverage
```


Полный функционал с описанием можно увидеть после ввода команды:

```shell
make help
```