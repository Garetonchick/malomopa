## RPS Test

Тестировали в условиях, где оба сервиса ограничены 1 CPU в деплое

```
    deploy:
      resources:
        limits:
          cpus: "1"
        reservations:
          cpus: "1"
```

С consistency = ONE получили:

```
Starting RPS test for handler assign_order. Setting up test.
Environment is ready.
Handler: assign_order, RPS: 531 requests/sec
Starting RPS test for handler acquire_order. Setting up test.
Environment is ready.
Handler: acquire_order, RPS: 147 requests/sec
Starting RPS test for handler cancel_order. Setting up test.
Environment is ready.
Handler: cancel_order, RPS: 134 requests/sec
RPS test completed.
```

Меняем consistency = ALL имеем:

```
Starting RPS test for handler assign_order. Setting up test.
Environment is ready.
Handler: assign_order, RPS: 460 requests/sec
Starting RPS test for handler acquire_order. Setting up test.
Environment is ready.
Handler: acquire_order, RPS: 103 requests/sec
Starting RPS test for handler cancel_order. Setting up test.
Environment is ready.
Handler: cancel_order, RPS: 132 requests/sec
RPS test completed.
```
