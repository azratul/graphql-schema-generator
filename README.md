# Schema Generator for GraphQL

## Description

Schemagen is a schema generator tool that converts Oracle, Postgres or MySQL Databases Schemas into GraphQL Types.

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

Oracle:

```zsh
graphql-schema-generator -motor="oracle" -schema="SCHEMA_OWNER" -dsn="DB_USER/DB_PASSWORD@(DESCRIPTION=(LOAD_BALANCE=ON)(FAILOVER=ON)(ADDRESS=(PROTOCOL=TCP)(HOST=DB_HOST)(PORT=DB_PORT))(CONNECT_DATA=(SERVER=DEDICATED)(SERVICE_NAME=db.orcl.com)))" -entities=TABLE1,TABLE2,TABLE3
```

Postgres:

```zsh
graphql-schema-generator -motor="postgres" -schema="SCHEMA" -dsn="postgres://PG_USER:PG_PASSWORD@DB_HOST/DB?sslmode=verify-full" -entities=TABLE1,TABLE2,TABLE3
```

MySQL:

```zsh
graphql-schema-generator -motor="mysql" -schema="SCHEMA" -dsn="DB_USER:DB_PASSWORD@TCP(127.0.0.1)/DB" -entities=TABLE1,TABLE2,TABLE3
```
