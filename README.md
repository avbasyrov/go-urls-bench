Сборка образа:
```bash
docker build -t urls-bench . 
```

Запуск:
```bash
docker run --name urls-bench -d -p 80:8080 urls-bench
```