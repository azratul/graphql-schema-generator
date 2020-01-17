# Schema Generator for GraphQL

## Description

Schemagen is a schema generator tool that converts Oracle Databases Schemas into GraphQL Types.

## Install

```zsh
go get github.com/azratul/graphql-schema-generator
go install github.com/azratul/graphql-schema-generator
```

## How to Use

```zsh
graphql-schema-generator -h
```

### Example:

```zsh
graphql-schema-generator -user="DB_USER" -password="DD_PASS" -dsn="(DESCRIPTION=(LOAD_BALANCE=ON)(FAILOVER=ON)(ADDRESS=(PROTOCOL=TCP)(HOST=DB_HOST)(PORT=DB_PORT))(CONNECT_DATA=(SERVER=DEDICATED)(SERVICE_NAME=db.orcl.com)))" -entities=TABLE1,TABLE2,TABLE3
```