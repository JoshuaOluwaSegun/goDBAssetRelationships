package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	apiLib "github.com/hornbill/goApiLib"
	hornbillHelpers "github.com/hornbill/goHornbillHelpers"
)

func main() {
	var err error
	//-- Start Time for Duration
	startTime = time.Now()
	//-- Start Time for Log File
	timeNow = time.Now().Format("20060102150405")
	logFileName = "assetRelationships" + timeNow + ".log"

	//-- Grab Flags
	flag.StringVar(&configFileName, "file", "conf.json", "Name of Configuration File To Load")
	flag.BoolVar(&configVersion, "version", false, "Return version and end")
	flag.BoolVar(&configDryrun, "dryrun", false, "Outputs the expected API calls to the log, without actually performing the API calls")
	flag.Parse()

	//-- If configVersion just output version number and die
	if configVersion {
		fmt.Printf("%v \n", version)
		return
	}

	//Load Config
	importConf = loadConfig()

	//Create shared espxmlmc session
	espXmlmc = apiLib.NewXmlmcInstance(importConf.InstanceID)
	espXmlmc.SetAPIKey(importConf.APIKey)
	//-- Output
	hornbillHelpers.Logger(2, "---- XMLMC Database Asset Relationship Import Utility V"+version+" ----", true, logFileName)
	hornbillHelpers.Logger(2, "Flag - Config File "+configFileName, true, logFileName)

	configManager, err = isConfigManagerInstalled()
	if err != nil {
		hornbillHelpers.Logger(4, "Error checking for Configuration Manager installation: "+err.Error(), true, logFileName)
	}

	if !configManager {
		hornbillHelpers.Logger(5, "Configuration Manager is not installed. Basic asset links ONLY will be imported", true, logFileName)
	} else {
		hornbillHelpers.Logger(2, "Configuration Manager has been detected. Basic asset links AND Configuration Manager dependencies/impacts will be imported", true, logFileName)
	}

	cacheHornbillRecords()

	//Get Asset Relationships from DB
	err = queryDatabase(false)
	if err != nil {
		os.Exit(1)
	}

	if importConf.RemoveLinks {
		//Get Asset Removal Relationships from DB
		err = queryDatabase(true)
		if err != nil {
			os.Exit(1)
		}
	}

	if len(assetRelationships) == 0 && len(assetDeleteRelationships) == 0 {
		hornbillHelpers.Logger(4, "No asset relationship or removal records returned from database queries", true, logFileName)
		os.Exit(1)
	}

	//Process Relationship Create/Update
	processRelationships()

	if importConf.RemoveLinks {
		//Process Relationship Removals
		processRelationshipRemovals()
	}

	//Output
	hornbillHelpers.Logger(2, "Processing Complete!", true, logFileName)
	hornbillHelpers.Logger(2, "* Relationship Records Found: "+strconv.Itoa(len(assetRelationships)), true, logFileName)
	hornbillHelpers.Logger(2, "* Asset Links Created: "+strconv.Itoa(counters.linksCreated), true, logFileName)
	hornbillHelpers.Logger(2, "* Asset Links Skipped (already exists): "+strconv.Itoa(counters.linksSkipped), true, logFileName)
	hornbillHelpers.Logger(2, "* Asset Links Failed: "+strconv.Itoa(counters.linksFailed), true, logFileName)
	hornbillHelpers.Logger(2, "* Dependency Records Created: "+strconv.Itoa(counters.depsCreated), true, logFileName)
	hornbillHelpers.Logger(2, "* Dependency Records Updated: "+strconv.Itoa(counters.depsUpdated), true, logFileName)
	hornbillHelpers.Logger(2, "* Dependency Records Skipped: "+strconv.Itoa(counters.depsSkipped), true, logFileName)
	hornbillHelpers.Logger(2, "* Dependency Records Failed: "+strconv.Itoa(counters.depsFailed), true, logFileName)
	hornbillHelpers.Logger(2, "* Dependency Records Update Failed: "+strconv.Itoa(counters.depsUpdateFailed), true, logFileName)
	hornbillHelpers.Logger(2, "* Impact Records Created: "+strconv.Itoa(counters.impsCreated), true, logFileName)
	hornbillHelpers.Logger(2, "* Impact Records Updated: "+strconv.Itoa(counters.impsUpdated), true, logFileName)
	hornbillHelpers.Logger(2, "* Impact Records Skipped: "+strconv.Itoa(counters.impsSkipped), true, logFileName)
	hornbillHelpers.Logger(2, "* Impact Records Failed: "+strconv.Itoa(counters.impsFailed), true, logFileName)
	hornbillHelpers.Logger(2, "* Impact Records Update Failed: "+strconv.Itoa(counters.impsUpdateFailed), true, logFileName)
	if importConf.RemoveLinks {
		hornbillHelpers.Logger(2, "* Remove Relationship Records Found: "+strconv.Itoa(len(assetDeleteRelationships)), true, logFileName)
		hornbillHelpers.Logger(2, "* Remove Asset Links Success: "+strconv.Itoa(counters.removeLinksSuccess), true, logFileName)
		hornbillHelpers.Logger(2, "* Remove Asset Links Skipped (doesn't exist): "+strconv.Itoa(counters.removeLinksSkipped), true, logFileName)
		hornbillHelpers.Logger(2, "* Remove Asset Links Failed: "+strconv.Itoa(counters.removeLinksFailed), true, logFileName)
		hornbillHelpers.Logger(2, "* Remove Dependency Records Success: "+strconv.Itoa(counters.removeDepsSuccess), true, logFileName)
		hornbillHelpers.Logger(2, "* Remove Dependency Records Skipped: "+strconv.Itoa(counters.removeDepsSkipped), true, logFileName)
		hornbillHelpers.Logger(2, "* Remove Dependency Records Failed: "+strconv.Itoa(counters.removeDepsFailed), true, logFileName)
		hornbillHelpers.Logger(2, "* Remove Impact Records Success: "+strconv.Itoa(counters.removeImpsSuccess), true, logFileName)
		hornbillHelpers.Logger(2, "* Remove Impact Records Skipped: "+strconv.Itoa(counters.removeImpsSkipped), true, logFileName)
		hornbillHelpers.Logger(2, "* Remove Impact Records Failed: "+strconv.Itoa(counters.removeImpsFailed), true, logFileName)
	}
}

func cacheHornbillRecords() {
	//Cache Service Manager Asset Records
	//-- Cache Assets first
	err := cacheAssets()
	if err != nil {
		hornbillHelpers.Logger(4, "Error when caching assets from Hornbill: "+err.Error(), true, logFileName)
		os.Exit(1)
	}

	//--Cache Links
	err = cacheAssetLinks()
	if err != nil {
		hornbillHelpers.Logger(4, "Error when caching asset links from Hornbill: "+err.Error(), true, logFileName)
		os.Exit(1)
	}

	if configManager {
		//Cache Config Manager Asset Records
		//--Cache Dependencies
		err = cacheAssetDependencies()
		if err != nil {
			hornbillHelpers.Logger(4, "Error when caching asset dependencies from Hornbill: "+err.Error(), true, logFileName)
			os.Exit(1)
		}

		//--Cache Impact Records
		err = cacheAssetImpacts()
		if err != nil {
			hornbillHelpers.Logger(4, "Error when caching asset impacts from Hornbill: "+err.Error(), true, logFileName)
			os.Exit(1)
		}
	}
}

//loadConfig -- Function to Load Configruation File
func loadConfig() sqlImportConfStruct {
	//-- Check Config File File Exists
	cwd, _ := os.Getwd()
	configurationFilePath := cwd + "/" + configFileName
	hornbillHelpers.Logger(1, "Loading Config File: "+configurationFilePath, false, logFileName)
	if _, fileCheckErr := os.Stat(configurationFilePath); os.IsNotExist(fileCheckErr) {
		hornbillHelpers.Logger(4, "No Configuration File", true, logFileName)
		os.Exit(102)
	}
	//-- Load Config File
	file, fileError := os.Open(configurationFilePath)
	//-- Check For Error Reading File
	if fileError != nil {
		hornbillHelpers.Logger(4, "Error Opening Configuration File: "+fmt.Sprintf("%v", fileError), true, logFileName)
	}

	//-- New Decoder
	decoder := json.NewDecoder(file)
	//-- New Var based on SQLimportConf
	esqlConf := sqlImportConfStruct{}
	//-- Decode JSON
	err := decoder.Decode(&esqlConf)
	//-- Error Checking
	if err != nil {
		hornbillHelpers.Logger(4, "Error Decoding Configuration File: "+fmt.Sprintf("%v", err), true, logFileName)
	}
	//-- Return New Congfig
	return esqlConf
}

func isConfigManagerInstalled() (bool, error) {
	xmlApps, err := espXmlmc.Invoke("session", "getApplicationList")
	if err != nil {
		hornbillHelpers.Logger(4, err.Error(), true, logFileName)
		return false, err
	}

	var xmlResponse methodCallResult
	err = xml.Unmarshal([]byte(xmlApps), &xmlResponse)
	if err != nil {
		hornbillHelpers.Logger(4, err.Error(), true, logFileName)
		return false, err
	}
	if xmlResponse.Status != "ok" {
		hornbillHelpers.Logger(4, xmlResponse.State.ErrorRet, true, logFileName)
		return false, errors.New(xmlResponse.State.ErrorRet)
	}
	for _, v := range xmlResponse.Params.Apps {
		if v.Name == "com.hornbill.configurationmanager" && v.Status == "installed" {
			return true, nil
		}
	}
	return false, err
}
