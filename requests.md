# Тестовые запросы для API pz3-http

1. Проверка здоровья

```powershell
curl.exe -i http://localhost:8080/health
```

2. Создание задачи

```powershell
curl.exe -i -X POST http://localhost:8080/tasks -H "Content-Type: application/json" -d '{"title":"Купить молоко"}'
```

3. Список задач

```powershell
curl.exe -i http://localhost:8080/tasks
```

4. Фильтр по title

```powershell
curl.exe -i "http://localhost:8080/tasks?q=молоко"
```

5. Получить задачу по id

```powershell
curl.exe -i http://localhost:8080/tasks/1
```

6. Ошибка валидации title

```powershell
curl.exe -i -X POST http://localhost:8080/tasks -H "Content-Type: application/json" -d '{"title":"ab"}'
```

7. Некорректный id

```powershell
curl.exe -i http://localhost:8080/tasks/abc
```

8. Несуществующая задача

```powershell
curl.exe -i http://localhost:8080/tasks/9999
```

Прочие запросы:

```
# PATCH отметить выполненной
curl.exe -i -X PATCH http://localhost:8080/tasks/1 -H "Content-Type: application/json" -d '{"done":true}'
# DELETE удалить задачу
curl.exe -i -X DELETE http://localhost:8080/tasks/1
```

## 10. Makefile и PowerShell-скрипты

### `Makefile`

```makefile
.PHONY: run build test

run:
	go run ./cmd/server

build:
	go build -o ./bin/server ./cmd/server

test:
	go test ./...
```