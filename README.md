# gokv
Key-value store in Go


# Functional goals

- Store arbitrary key-value pairs
- APIs for CRUD
- Persistance

# Features

- Idempotant

# HTTP endpoints
Might use swagger later.

Method|Endpoint|Purpose|Possible return types
--|--|--|--
Put a key-value pair|PUT|/api/v1/key/{key}|201, 500
Get the value given a key|GET|/api/v1/key/{key}|200, 404, 500
Delete a key-value pair|DELETE|/api/v1/key/{key}|200, 500

# Run tests

```sh
go test -cover ./...
```