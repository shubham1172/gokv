# gokv
Key-value store in Go


# Functional goals

- Store arbitrary key-value pairs
- APIs for CRUD
- Persistant (Failure resiliant)
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

# TODOs
- Configuration for buffer size, service address, log file name, etc.
- Dockerize
- Swagger
- More tests
- Makefile
- On startup, cleanup the logs
    - Essentially, remove each put-delete pair
    - Keep the latest overwrite for each put
    - Remove all other delete(s)
- Convert log to some binary format - protobuf? bson?