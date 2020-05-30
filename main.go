package main

import (
	"bytes"
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

type mysqlConnectionType struct {
	UserName string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
}

func newMySQLConnectionType() mysqlConnectionType {
	return mysqlConnectionType{
		UserName: "root",
		Password: "password",
		Host:     "127.0.0.1",
		Port:     3306,
	}
}

func (connectionConfig mysqlConnectionType) dataSourceName() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)",
		connectionConfig.UserName,
		connectionConfig.Password,
		connectionConfig.Host,
		connectionConfig.Port)
}

type configType struct {
	MySQLConnection mysqlConnectionType `json:"connection"`
}

func newConfigType() configType {
	return configType{
		MySQLConnection: newMySQLConnectionType(),
	}
}

func readConfig(filePath string) (ret configType, err error) {
	ret = newConfigType()
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

const (
	generateConfigJSONCommand = "generate-config-json"
	helpCommand               = "help"
	runCommand                = "run"
)

func helpPage() error {
	return fmt.Errorf(generateConfigJSONCommand + " " + runCommand + " " + helpCommand)
}

func main() {
	flag.ErrHelp = helpPage()
	flag.Usage = func() { fmt.Println(helpPage().Error()) }

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	runCommandFlagSet := flag.NewFlagSet(runCommand, flag.ExitOnError)
	configPath := runCommandFlagSet.String("config", "", "path of config file")
	port := runCommandFlagSet.Int("port", 0, "port")

	switch os.Args[1] {
	case generateConfigJSONCommand:
		configJSON, err := json.Marshal(newConfigType())
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		out := new(bytes.Buffer)
		json.Indent(out, configJSON, "", "  ")
		fmt.Println(out.String())
		os.Exit(0)
	case helpCommand:
		fmt.Println(helpPage())
		os.Exit(0)
	case runCommand:
		runCommandFlagSet.Parse(os.Args[2:])

		config, err := readConfig(*configPath)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		if *port != 0 {
			config.MySQLConnection.Port = *port
		}

		connectionInformation := config.MySQLConnection.dataSourceName()
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
	}
	return
}
