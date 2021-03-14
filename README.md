# gokv
Key-value store in Go


# Functional goals

- Store arbitrary key-value pairs
- APIs for CRUD
- Idempotent PUT/DELETE
- Persistent (Failure resilient)
    - Transactional logs 
        - Seq no, Event type, key, value

# HTTP endpoints

Method|Endpoint|Purpose|Possible return types
--|--|--|--
Put a key-value pair|PUT|/api/v1/key/{key}|201, 400, 500
Get the value given a key|GET|/api/v1/key/{key}|200, 400, 404, 500
Delete a key-value pair|DELETE|/api/v1/key/{key}|200, 400, 500

# Handy commands

```sh
go test -cover .\...
go fmt .\...
godoc -http=:8081
docker run --rm --name pgdb -p 5432:5432 -e POSTGRES_PASSWORD=password -d postgres
```

# TODOs
- **Convert TODOs to GitHub issues**
- Configuration for buffer size, service address, log file name, etc.
- Dockerfile/compose for prod
- Find hot-reloading alternative for windows
    - fsnotify refuses to work on windows containers
- Swagger
- Refactor logging
- More tests
- Makefile
- On startup, cleanup the logs
    - Essentially, remove each put-delete pair
    - Keep the latest overwrite for each put
    - Remove all other delete(s)
- Encode whitespaces/linebreaks in key/value for logging
- Convert file logger to some binary format - protobuf? bson?
- Use contexts
- Authentication