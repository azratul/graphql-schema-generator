package main

import (
	"database/sql"
	"flag"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

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

func init() {
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
	var data2 string
	var type_query string
	var type_mutation string
	var query string
	var scalar bool

	var re = regexp.MustCompile(`: (.*)`)

	if *motor == "godror" {
		// oracle
		query = `SELECT COLUMN_NAME, DATA_TYPE, DATA_SCALE, NULLABLE FROM ALL_TAB_COLUMNS WHERE UPPER(TABLE_NAME)=UPPER(:1) AND UPPER(OWNER)=UPPER(:2) ORDER BY COLUMN_ID`
	} else {
		// postgres or mysql
		bind := [2]string{"?", "?"}
		if *motor == "postgres" {
			bind = [2]string{"$1", "$2"}
		}

		query = `SELECT COLUMN_NAME, DATA_TYPE, NUMERIC_SCALE, IS_NULLABLE FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME=` + bind[0] + ` AND TABLE_SCHEMA=` + bind[1] + ``
	}

	stm, err := db.Prepare(query)

	if err != nil {
		log.Println(err)
	}

	for _, x := range entities {
		entity := strings.TrimSpace(x)
		entityTitle := strings.Title(strings.ToLower(entity))
		rows, err := stm.Query(entity, *schema)
		if err != nil {
			log.Fatalf("Query error: %s", err)
		}
		defer rows.Close()

		type_query += "    getRow" + entityTitle + "(filter: Filter" + entityTitle + "): " + entityTitle + "\n"
		type_query += "    getRows" + entityTitle + "(filter: FilterAll" + entityTitle + ", pagination: Pagination): [" + entityTitle + "]\n"
		type_mutation += "    create" + entityTitle + "(input: In" + entityTitle + "): " + entityTitle + "\n"
		type_mutation += "    update" + entityTitle + "(input: Filter" + entityTitle + ", filter: Filter" + entityTitle + "): " + entityTitle + "\n"

		data += "type " + entityTitle + " {\n"
		for rows.Next() {
			var column_name string
			var data_type string
			var data_scale []byte
			var nullable string
			if err := rows.Scan(&column_name, &data_type, &data_scale, &nullable); err != nil {
				log.Fatalf("Scan error: %s", err)
			}

			data_type = strings.ToUpper(data_type)

			if data_type == "VARCHAR" ||
				data_type == "VARCHAR2" ||
				data_type == "NVARCHAR" ||
				data_type == "NVARCHAR2" ||
				data_type == "CHAR" ||
				data_type == "TEXT" {
				data_type = "String"
			} else if data_type == "DATE" ||
				strings.Contains(data_type, "TIMESTAMP") ||
				data_type == "DATETIME" {
				data_type = "Time"
				scalar = true
			} else if data_type == "BOOLEAN" {
				data_type = "Boolean"
			} else {
				i, _ := strconv.Atoi(string(data_scale))

				if i > 0 {
					data_type = "Float"
				} else {
					data_type = "Int"
				}
			}

			if nullable == "N" || nullable == "NO" {
				data_type += "!"
			}

			data += "    " + strings.ToLower(column_name) + ": " + data_type + "\n"
		}
		data += "}\n\n"
	}
	data2 = strings.Replace(data, "type ", "input In", -1)
	data += data2
	data2 = strings.Replace(data2, "input In", "input Filter", -1)
	data2 = strings.Replace(data2, "!", "", -1)
	data += data2
	data2 = strings.Replace(data2, "input Filter", "input FilterAll", -1)
	data2 = re.ReplaceAllString(data2, `: [$1]`)
	data += data2

	data += "input Pagination {\n\tpageNumber: Int!\n\tpageSize: Int!\n}\n\n"

	data += "type Query {\n" + type_query + "}\n\n"
	data += "type Mutation {\n" + type_mutation + "}\n\n"

	if scalar {
		data += "scalar Time"
	}

	return data
}

func Write(data string) {
	f, err := os.OpenFile("schema.graphqls", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)

	if err != nil {
		log.Fatalf("Opening file error: %s", err)
	}

	_ = f.Truncate(0)
	_, _ = f.Seek(0, 0)

	if _, err = f.WriteString(data); err != nil {
		log.Fatalf("Writing error!: %s", err)
	}
}
