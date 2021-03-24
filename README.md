# gokv
Key-value store in Go

# What is it?

- Gokv is a key-value store with REST APIs for adding/deleting/fetching key-value pairs
- All requests are idempotent
- Gokv is failure resilient. It uses transaction logs to store add & delete events and rollbacks each time it starts
  - Supports file based and db based transaction logging
- Gokv is configurable using a config file or environment variables

# HTTP endpoints

Purpose|Method|Endpoint|Possible return types
--|--|--|--
Put a key-value pair|PUT|/api/v1/key/{key}|201, 400, 500
Get the value given a key|GET|/api/v1/key/{key}|200, 400, 404, 500
Delete a key-value pair|DELETE|/api/v1/key/{key}|200, 400, 500

# Configuring gokv

Note, if environment variables are set, they will override the configuration file `config.yml`. 

## Supported configuration

|config.yml|environment|purpose|default
--|--|--|--
server.address|GOKV_SERVER_ADDRESS|Server hosting address including port number. Example: "0.0.0.0:8080"|":8000"
logging.logtype|GOKV_LOGGING_LOGTYPE|Type of logging mechanism to use. Can be "file" or "database" (pg)|"file"
logging.logfilename|GOKV_LOGGING_LOGFILENAME|Name of the file to write logs to|"transactions.log"
database.dbname|GOKV_DATABASE_DBNAME|Database name|"postgres"
database.host|GOKV_DATABASE_HOST|Database host|"postgres"
database.user|GOKV_DATABASE_USER|Database username|"postgres"
database.password|GOKV_DATABASE_PASSWORD|Database password|"password"
database.sslstatus|GOKV_DATABASE_SSLSTATUS|Database SSL status. Can be "require" or "disable"|"disable"

<br/>

Note, 
1. GOKV_DATABASE_* or database.* configuration is only relevant if logging type is set to "database"
1. GOKV_LOGGING_LOGFILENAME or logging.logfilename is only relevant if logging type is set to "file"


# Handy commands

```sh
go test -cover .\...
go fmt .\...
godoc -http=:8081
docker run --rm --name pgdb -p 5432:5432 -e POSTGRES_PASSWORD=password -d postgres
```

# TODOs
- **Convert TODOs to GitHub issues**
- Clean up the README
- CLI support for interacting with store
- Dockerfile/compose for prod
- Find hot-reloading alternative for windows
    - fsnotify refuses to work on windows containers
- Swagger
- TLS
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
