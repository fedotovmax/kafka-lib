Download:

go get github.com/fedotovmax/kafka-lib@v1.0.13

# 1. Для корректного использования в проекте нужно добавить в миграцию следующую таблицу:



## Далее создать все сущности:

### пакет adapters/db/postgres/- создать адаптер для postgresql

### пакет event_sender - создать отправителя событий. Требует storage (adapter postgres)

### пакет kafka создать producer и consumer (по требованию)

### создать финальный outbox processor из пакета outbox, передав все требуемые ему завивимости.

# 2. Или сделать свою реализацию, но требуется реализовать интерфейсы:

{ is a shell keyword
