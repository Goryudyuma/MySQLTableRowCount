package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

type tableNameType struct {
	Schema string `json:"schema"`
	Name   string `json:"name"`
}

func tableNames(connectionInformation string) (ret []tableNameType, err error) {
	informationSchemaDB, err := sql.Open("mysql", fmt.Sprintf("%s/information_schema", connectionInformation))
	if err != nil {
		return
	}
	defer informationSchemaDB.Close()

	rows, err := informationSchemaDB.Query("SELECT TABLE_SCHEMA, TABLE_NAME FROM tables")
	for rows.Next() {
		var tableSchema, tableName string
		err = rows.Scan(&tableSchema, &tableName)
		if err != nil {
			return
		}

		ret = append(ret, tableNameType{Schema: tableSchema, Name: tableName})
	}

	return
}

type tableInfoType struct {
	tableNameType
	Num int64 `json:"num"`
}

func tableInfo(connectionInformation string, tableNameList []tableNameType) (ret []tableInfoType, err error) {
	for _, tableName := range tableNameList {
		db, err := sql.Open("mysql", fmt.Sprintf("%s/%s", connectionInformation, tableName.Schema))
		if err != nil {
			return nil, err
		}

		defer db.Close()
		var num int64
		if err = db.QueryRow("SELECT count(1) FROM " + tableName.Name).Scan(&num); err != nil {
			return nil, err
		}

		ret = append(ret, tableInfoType{tableNameType: tableName, Num: num})
	}
	return
}

func main() {
	var userName, password, host string
	userName = "root"
	password = "mysql123"
	host = "127.0.0.1"
	connectionInformation := fmt.Sprintf("%s:%s@tcp(%s:3306)", userName, password, host)
	tableNameList, err := tableNames(connectionInformation)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	tableInfoList, err := tableInfo(connectionInformation, tableNameList)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	json, err := json.Marshal(tableInfoList)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Println(string(json))

	return
}
