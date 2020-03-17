package main

import (
    "flag"
    "log"
    "os"
    "strings"
    "strconv"
    "database/sql"

    //Oracle
    _ "github.com/godror/godror"
    //Postgres
    _ "github.com/lib/pq"
    //MySQL
    _ "github.com/go-sql-driver/mysql"
)

var dsn *string
var entities *string
var schema *string
var motor *string

func init(){
    dsn = flag.String("dsn", "", "Data source name\nEx:\n\t-dsn=\"{USER}/{PASSWORD}@(DESCRIPTION=(LOAD_BALANCE=ON)(FAILOVER=ON)(ADDRESS=(PROTOCOL={PROTOCOL})(HOST={HOST})(PORT={PORT}))(CONNECT_DATA=(SERVER=DEDICATED)(SERVICE_NAME={SERVICE_NAME})))\"")
    entities = flag.String("entities", "", "Array of entities(comma separated)\nEx:\n\t-entities=table1,table2,table3")
    motor = flag.String("motor", "", "Database motor\nEx:\n\t-motor=\"oracle\"")
    schema = flag.String("schema", "", "Database schema (For oracle you should use the owner of the schema)\nEx:\n\t-schema=\"SCHEMA_OWNER\"")

    flag.Parse()

    if *dsn == "" ||
       *entities == "" ||
       *motor == "" {
        log.Println("DSN, Entities or Motor aren't defined!")
        os.Exit(2)
    }

    if *motor == "oracle" {
        *motor = "godror"
    }
}

func main() {
    db, err := sql.Open(*motor, *dsn)

    if err != nil {
        log.Fatalf("Error: %s", err)
    }

    enttsarray := strings.Split(*entities, ",")

    *schema = strings.TrimSpace(*schema)

    data := makeSchemas(db, enttsarray)
    Write(data)
}

func makeSchemas(db *sql.DB, entities []string) string {
    var data string
    var query string
    var querySelect  string
    var queryInsert  string
    var queryInsert2 string
    var queryUpdate  string
    var definition   string

    if *motor == "godror" {
        // oracle
        query = `SELECT COLUMN_NAME, DATA_TYPE, DATA_SCALE FROM ALL_TAB_COLUMNS WHERE UPPER(TABLE_NAME)=UPPER(:1) AND UPPER(OWNER)=UPPER(:2) ORDER BY COLUMN_ID`
    } else {
        // postgres or mysql
        bind := [2]string{"?", "?"}
        if *motor == "postgres" {
            bind = [2]string{"$1", "$2"}
        }

        query = `SELECT COLUMN_NAME, DATA_TYPE, NUMERIC_SCALE FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME=` + bind[0] + ` AND TABLE_SCHEMA=` + bind[1] + ``;
    }

    stm, err := db.Prepare(query)

    if err != nil {
        log.Println(err)
    }

    for _, x := range entities {
        entity := strings.TrimSpace(x)
        rows, err := stm.Query(entity, *schema)
        if err != nil {
            log.Fatalf("Query error: %s",err)
        }
        defer rows.Close()

	querySelect  = entity + ":SELECT "
	queryInsert  = entity + ":INSERT INTO " + entity + "("
	queryInsert2 = ""
	queryUpdate  = entity + ":UPDATE " + entity + " SET "
	definition   = entity + ":DEFINITION:"
        for rows.Next() {
            var column_name string
            var data_type string
            var data_scale []byte
            if err := rows.Scan(&column_name, &data_type, &data_scale); err != nil {
                log.Fatalf("Scan error: %s",err)
            }

            data_type = strings.ToUpper(data_type)

            if  data_type == "VARCHAR" ||
                data_type == "VARCHAR2" ||
                data_type == "CHAR" ||
                data_type == "TEXT" {
                data_type = "string"
            } else if data_type == "DATE" ||
                data_type == "DATETIME" {
                data_type = "time"
            } else if data_type == "BOOLEAN" {
                data_type = "bool"
            } else {
                i, _ := strconv.Atoi(string(data_scale))

                if i > 0 {
                    data_type = "float"
                } else {
                    data_type = "int"
                }
            }

	    querySelect  += column_name + ","
	    queryInsert2 += ":" + column_name + ","
	    queryUpdate  += column_name + " = :" + strings.ToLower(column_name) + ","
	    title := ""

            if strings.ToUpper(column_name) == "ID" {
		title = strings.ToUpper(column_name)
            } else {
	        title = strings.Replace(strings.ToLower(column_name),"_"," ",-1)
	        title = strings.Title(title)
	        title = strings.Replace(title," ","",-1)
            }

	    definition += title + "," + column_name + "," + data_type + ";"
        }
	querySelect  = strings.TrimRight(querySelect, ",")
	queryInsert2 = strings.TrimRight(queryInsert2, ",")
	queryInsert += strings.Replace(queryInsert2, ":", "", -1)
	queryUpdate  = strings.TrimRight(queryUpdate, ",")
	definition   = strings.TrimRight(definition, ";")
        data += querySelect + " FROM " + entity + " WHERE 1=1\n" + queryInsert + ") VALUES (" + strings.ToLower(queryInsert2) + ")\n" + queryUpdate + " WHERE 1=1\n" + definition + "\n\n"
    }

    return data
}

func Write(data string) {
    f, err := os.OpenFile("queries.txt", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)

    if err != nil {
        log.Fatalf("Opening file error: %s", err)
    }

    _ = f.Truncate(0)
    _, _ = f.Seek(0,0)

    if _, err = f.WriteString(data); err != nil {
        log.Fatalf("Writing error!: %s", err)
    }
}
