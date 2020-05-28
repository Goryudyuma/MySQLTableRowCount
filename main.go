package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
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

	err = informationSchemaDB.Ping()
	if err != nil {
		return
	}

	rows, err := informationSchemaDB.Query("SELECT TABLE_SCHEMA, TABLE_NAME FROM tables WHERE TABLE_SCHEMA = 'test'")
	if err != nil {
		return
	}

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

type mysqlConnection struct {
	UserName string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
}

type configType struct {
	MySQLConnection mysqlConnection `json:"connection"`
}

func readConfig(filePath string) (ret configType, err error) {
	if filePath == "" {
		return
	}

	configFileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return
	}

	err = json.Unmarshal(configFileBytes, &ret)
	return
}

func main() {

	configPath := flag.String("config", "", "path of config file")
	port := flag.Int("port", 0, "port")
	flag.Parse()
	config, err := readConfig(*configPath)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if *port != 0 {
		config.MySQLConnection.Port = *port
	}

	connectionInformation := fmt.Sprintf("%s:%s@tcp(%s:%d)", config.MySQLConnection.UserName, config.MySQLConnection.Password, config.MySQLConnection.Host, config.MySQLConnection.Port)
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
