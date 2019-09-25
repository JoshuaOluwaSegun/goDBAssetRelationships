package main

import (
	"errors"
	"fmt"
	"strconv"

	//SQL Package
	hornbillHelpers "github.com/hornbill/goHornbillHelpers"
	"github.com/jmoiron/sqlx"

	//SQL Drivers
	_ "github.com/alexbrainman/odbc" //ODBC Driver
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/hornbill/mysql320" //MySQL v3.2.0 to v5 driver - Provides SWSQL (MySQL 4.0.16) support
)

//buildConnectionString -- Build the connection string for the SQL driver
func buildConnectionString() string {
	if importConf.DBConf.Database == "" ||
		importConf.DBConf.Authentication == "SQL" && (importConf.DBConf.UserName == "" || importConf.DBConf.Password == "") {
		//Conf not set - log error and return empty string
		hornbillHelpers.Logger(4, "Database configuration not set.", true, logFileName)
		return ""
	}
	if importConf.DBConf.Driver != "odbc" {
		hornbillHelpers.Logger(1, "Connecting to Database Server: "+importConf.DBConf.Server, true, logFileName)
	} else {
		hornbillHelpers.Logger(1, "Connecting to ODBC Data Source: "+importConf.DBConf.Database, true, logFileName)
	}

	connectString := ""
	switch importConf.DBConf.Driver {
	case "mssql":
		connectString = "server=" + importConf.DBConf.Server
		connectString = connectString + ";database=" + importConf.DBConf.Database
		if importConf.DBConf.Authentication == "Windows" {
			connectString = connectString + ";Trusted_Connection=True"
		} else {
			connectString = connectString + ";user id=" + importConf.DBConf.UserName
			connectString = connectString + ";password=" + importConf.DBConf.Password
		}

		if !importConf.DBConf.Encrypt {
			connectString = connectString + ";encrypt=disable"
		}
		if importConf.DBConf.Port != 0 {
			dbPortSetting := strconv.Itoa(importConf.DBConf.Port)
			connectString = connectString + ";port=" + dbPortSetting
		}
	case "mysql":
		connectString = importConf.DBConf.UserName + ":" + importConf.DBConf.Password
		connectString = connectString + "@tcp(" + importConf.DBConf.Server + ":"
		if importConf.DBConf.Port != 0 {
			dbPortSetting := strconv.Itoa(importConf.DBConf.Port)
			connectString = connectString + dbPortSetting
		} else {
			connectString = connectString + "3306"
		}
		connectString = connectString + ")/" + importConf.DBConf.Database
	case "mysql320":
		dbPortSetting := "3306"
		if importConf.DBConf.Port != 0 {
			dbPortSetting = strconv.Itoa(importConf.DBConf.Port)
		}
		connectString = "tcp:" + importConf.DBConf.Server + ":" + dbPortSetting
		connectString = connectString + "*" + importConf.DBConf.Database + "/" + importConf.DBConf.UserName + "/" + importConf.DBConf.Password
	case "odbc":
		connectString = "DSN=" + importConf.DBConf.Database + ";UID=" + importConf.DBConf.UserName + ";PWD=" + importConf.DBConf.Password
	}
	return connectString
}

//queryDatabase -- Query Asset Relationships Database
func queryDatabase() error {
	connString := buildConnectionString()
	if connString == "" {
		hornbillHelpers.Logger(4, " [DATABASE] Database Connection String Empty. Check the DBConf section of your configuration.", true, logFileName)
		return errors.New("database connection string empty - check the dbconf section of your configuration")
	}
	//Connect to the JSON specified DB
	db, err := sqlx.Open(importConf.DBConf.Driver, connString)
	if err != nil {
		hornbillHelpers.Logger(4, " [DATABASE] Database Connection Error: "+fmt.Sprintf("%v", err), true, logFileName)
		return err
	}
	defer db.Close()
	//Check connection is open
	err = db.Ping()
	if err != nil {
		hornbillHelpers.Logger(4, " [DATABASE] [PING] Database Ping Error: "+fmt.Sprintf("%v", err), true, logFileName)
		return err
	}
	hornbillHelpers.Logger(3, "[DATABASE] Connection Successful", true, logFileName)
	hornbillHelpers.Logger(3, "[DATABASE] Running database query for asset relationships. Please wait...", true, logFileName)
	hornbillHelpers.Logger(3, "[DATABASE] Query: "+importConf.Query, false, logFileName)
	//Run Query
	rows, err := db.Queryx(importConf.Query)
	if err != nil {
		hornbillHelpers.Logger(4, " [DATABASE] Database Query Error: "+fmt.Sprintf("%v", err), true, logFileName)
		return err
	}
	defer rows.Close()

	//Build map full of asset relationship records
	intAssetCount := 0
	intAssetSuccess := 0
	for rows.Next() {
		intAssetCount++
		results := make(map[string]interface{})
		err = rows.MapScan(results)
		if err != nil {
			hornbillHelpers.Logger(4, " [DATABASE] Data Unmarshal Error: "+fmt.Sprintf("%v", err), true, logFileName)
		} else {
			//Stick marshalled data map in to parent slice
			assetRelationships = append(assetRelationships, results)
			intAssetSuccess++
		}
	}
	hornbillHelpers.Logger(3, "[DATABASE] "+strconv.Itoa(intAssetSuccess)+" of "+strconv.Itoa(intAssetCount)+" asset relationship records successfully retrieved ready for processing.", true, logFileName)
	return nil
}
