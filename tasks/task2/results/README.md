при запуске сервиса первый раз все необходимые таблицы создадутся, если не были ранее созданы
речь идет о таблице `booking` в новом сервисе `booking-service`

для kafka необходимо создать топик, примерный синтаксис
```shell
  docker run -it --rm \
    --link kafka-broker \
    confluentinc/cp-kafka:7.2.1 \
    kafka-topics --create \
    --topic bookings \
    --partitions 3 \
    --replication-factor 1 \
    --bootstrap-server kafka:9092
```

