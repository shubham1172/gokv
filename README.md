# gokv
Key-value store in Go


# Functional goals

- Store arbitrary key-value pairs
- APIs for CRUD
- Persistance
    - Transactional logs - TODO: binary instead of plain-text
        - Seq no, Event type, key, value

# Features

- Idempotant

# HTTP endpoints

Method|Endpoint|Purpose|Possible return types
--|--|--|--
Put a key-value pair|PUT|/api/v1/key/{key}|201, 500
Get the value given a key|GET|/api/v1/key/{key}|200, 404, 500
Delete a key-value pair|DELETE|/api/v1/key/{key}|200, 500

# Handy commands

```sh
go test -cover .\...
go fmt .\...
godoc -http=:8081
```

# Other TODOs
- Configuration for buffer size, service address, log file name, etc.
- Dockerize
- Swagger
- More tests
- Makefile