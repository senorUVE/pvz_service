## Важно

Интеграционное тестирование было доделано после дедлайна (в течение 4 часов после его окончания)

Планируется доделать дополнительные задания

коммиты, содержащие их, содержат префикс в названии: `after deadline`

Мной было принято решение предоставить реализацию пунктов ТЗ, несмотря на окончание дедлайна.

## Конфигурация проекта

Для демонстрации работы используются тестовые конфигурационные файлы:
`deploy/.env`
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

Генерация dto endpoint'ов по openapi схеме:
```shell
tasl gen-dto
```


Полный функционал с описанием можно увидеть после ввода команды:

```shell
task help
```
## Интеграционное тестирование
В директории test, там же конфиг test_config.yaml

для проведения тестов также была поднята тестовая БД (запускать нужно из корня проекта, чтобы был доступ к миграциям):

```shell
docker run --name db -p 5431:5432 \
    -e POSTGRES_USER=postgres \
    -e POSTGRES_PASSWORD=1234 \
    -e POSTGRES_DB=mydb \
    -v ./migrations:/docker-entrypoint-initdb.d \
    -d postgres:17
```

запуск тестов:

```shell
go test ./test 
```