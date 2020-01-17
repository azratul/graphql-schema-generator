package main

import (
    "flag"
    "log"
    "os"
    "strings"
    "database/sql"

    _ "github.com/godror/godror"
)

var dsn *string
var entities *string
var user *string
var password *string

func init(){
    dsn = flag.String("dsn", "", "Data source name\nEx:\n\tgraphql-schema-generator -dsn=\"(DESCRIPTION=(LOAD_BALANCE=ON)(FAILOVER=ON)(ADDRESS=(PROTOCOL={PROTOCOL})(HOST={HOST})(PORT={PORT}))(CONNECT_DATA=(SERVER=DEDICATED)(SERVICE_NAME={SERVICE_NAME})))\"")
    entities = flag.String("entities", "", "Array of entities(comma separated)\nEx:\n\tgraphql-schema-generator -entities=table1,table2,table3")
    user = flag.String("user", "", "Database user\nEx:\n\tgraphql-schema-generator -user=\"DB_USER\"")
    password = flag.String("password", "", "Database password\nEx:\n\tgraphql-schema-generator -password=\"PASSWORD1234\"")

    flag.Parse()

    if *dsn == "" ||
       *entities == "" ||
       *user == "" ||
       *password == "" {
        log.Println("DSN or Entities aren't defined!")
        os.Exit(2)
    }
}

func main() {
    db, err := sql.Open("godror", *user + "/" + *password + *dsn)

    if err != nil {
        log.Fatalf("Error: %s", err)
    }

    enttsarray := strings.Split(*entities, ",")

    *user = strings.TrimSpace(*user)

    data := makeSchemas(db, enttsarray)
    Write(data)
}

func makeSchemas(db *sql.DB, entities []string) string {
    var data string
    stm, err := db.Prepare(`SELECT COLUMN_NAME, DATA_TYPE, NULLABLE FROM ALL_TAB_COLUMNS WHERE TABLE_NAME=:1 AND OWNER=:2 ORDER BY COLUMN_ID`)

    if err != nil {
        log.Println(err)
    }

    for _, x := range entities {
        entity := strings.TrimSpace(x)
        rows, err := stm.Query(entity, *user)
        if err != nil {
            log.Fatalf("Query error: %s",err)
        }
        defer rows.Close()

        data += "type " + strings.Title(strings.ToLower(entity)) + " {\n"
        for rows.Next() {
            var column_name string
            var data_type string
            var nullable string
            if err := rows.Scan(&column_name, &data_type, &nullable); err != nil {
                log.Fatalf("Scan error: %s",err)
            }

            data_type = strings.ToUpper(data_type)

            if  data_type == "VARCHAR" ||
                data_type == "VARCHAR2" ||
                data_type == "CHAR" ||
                data_type == "TEXT" {
                data_type = "String"
            } else {
                data_type = "Int"
            }

            if nullable == "N" {
                data_type += "!"
            }

            column_name = strings.Title(strings.ToLower(column_name))

            data += "    " + column_name + ": " + data_type + "\n"
        }
        data += "}\n\n"
    }

    return data
}

func Write(data string) {
    f, err := os.OpenFile("schema.graphql", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)

    if err != nil {
        log.Fatalf("Opening file error: %s", err)
    }

    _ = f.Truncate(0)
    _, _ = f.Seek(0,0)

    if _, err = f.WriteString(data); err != nil {
        log.Fatalf("Writing error!: %s", err)
    }
}