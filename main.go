package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	apiLib "github.com/hornbill/goApiLib"
	hornbillHelpers "github.com/hornbill/goHornbillHelpers"
	"github.com/tcnksm/go-latest"
)

func main() {
	var err error

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

	checkVersion()
	logger(2, "---- XMLMC Database Asset Relationship Import Utility V"+version+" ----", true, true)
	logger(2, "Flag - Config File "+configFileName, true, true)

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
		logger(4, "No asset relationship or removal records returned from database queries", true, true)
		os.Exit(1)
	}

	//Process Relationship Create/Update
	processRelationships()

	if importConf.RemoveLinks {
		//Process Relationship Removals
		processRelationshipRemovals()
	}

	//Output
	logger(2, "Processing Complete!", true, true)
	logger(2, "* Relationship Records Found: "+strconv.Itoa(len(assetRelationships)), true, true)
	logger(2, "* Asset Links Created: "+strconv.Itoa(counters.linksCreated), true, true)
	logger(2, "* Asset Links Skipped (already exists): "+strconv.Itoa(counters.linksSkipped), true, true)
	logger(2, "* Asset Links Failed: "+strconv.Itoa(counters.linksFailed), true, true)
	logger(2, "* Dependency Records Created: "+strconv.Itoa(counters.depsCreated), true, true)
	logger(2, "* Dependency Records Updated: "+strconv.Itoa(counters.depsUpdated), true, true)
	logger(2, "* Dependency Records Skipped: "+strconv.Itoa(counters.depsSkipped), true, true)
	logger(2, "* Dependency Records Failed: "+strconv.Itoa(counters.depsFailed), true, true)
	logger(2, "* Dependency Records Update Failed: "+strconv.Itoa(counters.depsUpdateFailed), true, true)
	logger(2, "* Impact Records Created: "+strconv.Itoa(counters.impsCreated), true, true)
	logger(2, "* Impact Records Updated: "+strconv.Itoa(counters.impsUpdated), true, true)
	logger(2, "* Impact Records Skipped: "+strconv.Itoa(counters.impsSkipped), true, true)
	logger(2, "* Impact Records Failed: "+strconv.Itoa(counters.impsFailed), true, true)
	logger(2, "* Impact Records Update Failed: "+strconv.Itoa(counters.impsUpdateFailed), true, true)
	if importConf.RemoveLinks {
		logger(2, "* Remove Relationship Records Found: "+strconv.Itoa(len(assetDeleteRelationships)), true, true)
		logger(2, "* Remove Asset Links Success: "+strconv.Itoa(counters.removeLinksSuccess), true, true)
		logger(2, "* Remove Asset Links Skipped (doesn't exist): "+strconv.Itoa(counters.removeLinksSkipped), true, true)
		logger(2, "* Remove Asset Links Failed: "+strconv.Itoa(counters.removeLinksFailed), true, true)
		logger(2, "* Remove Dependency Records Success: "+strconv.Itoa(counters.removeDepsSuccess), true, true)
		logger(2, "* Remove Dependency Records Skipped: "+strconv.Itoa(counters.removeDepsSkipped), true, true)
		logger(2, "* Remove Dependency Records Failed: "+strconv.Itoa(counters.removeDepsFailed), true, true)
		logger(2, "* Remove Impact Records Success: "+strconv.Itoa(counters.removeImpsSuccess), true, true)
		logger(2, "* Remove Impact Records Skipped: "+strconv.Itoa(counters.removeImpsSkipped), true, true)
		logger(2, "* Remove Impact Records Failed: "+strconv.Itoa(counters.removeImpsFailed), true, true)
	}
}

func cacheHornbillRecords() {
	//Cache Service Manager Asset Records
	//-- Cache Assets first
	err := cacheAssets()
	if err != nil {
		logger(4, "Error when caching assets from Hornbill: "+err.Error(), true, true)
		os.Exit(1)
	}

	//--Cache Links
	err = cacheAssetLinks()
	if err != nil {
		logger(4, "Error when caching asset links from Hornbill: "+err.Error(), true, true)
		os.Exit(1)
	}

	//Cache Config Manager Asset Records
	//--Cache Dependencies
	err = cacheAssetDependencies()
	if err != nil {
		logger(4, "Error when caching asset dependencies from Hornbill: "+err.Error(), true, true)
		os.Exit(1)
	}

	//--Cache Impact Records
	err = cacheAssetImpacts()
	if err != nil {
		logger(4, "Error when caching asset impacts from Hornbill: "+err.Error(), true, true)
		os.Exit(1)
	}
}

//loadConfig -- Function to Load Configruation File
func loadConfig() sqlImportConfStruct {
	//-- Check Config File File Exists
	cwd, _ := os.Getwd()
	configurationFilePath := cwd + "/" + configFileName
	logger(1, "Loading Config File: "+configurationFilePath, false, false)
	if _, fileCheckErr := os.Stat(configurationFilePath); os.IsNotExist(fileCheckErr) {
		logger(4, "No Configuration File", true, true)
		os.Exit(102)
	}
	//-- Load Config File
	file, fileError := os.Open(configurationFilePath)
	//-- Check For Error Reading File
	if fileError != nil {
		logger(4, "Error Opening Configuration File: "+fmt.Sprintf("%v", fileError), true, false)
	}

	//-- New Decoder
	decoder := json.NewDecoder(file)
	//-- New Var based on SQLimportConf
	esqlConf := sqlImportConfStruct{}
	//-- Decode JSON
	err := decoder.Decode(&esqlConf)
	//-- Error Checking
	if err != nil {
		logger(4, "Error Decoding Configuration File: "+fmt.Sprintf("%v", err), true, false)
	}
	//-- Return New Congfig
	return esqlConf
}

//-- Check Latest
func checkVersion() {
	githubTag := &latest.GithubTag{
		Owner:      "hornbill",
		Repository: appName,
	}

	res, err := latest.Check(githubTag, version)
	if err != nil {
		msg := "Unable to check utility version against Github repository: " + err.Error()
		logger(4, msg, true, true)
		return
	}
	if res.Outdated {
		msg := "v" + version + " is not latest, you should upgrade to " + res.Current + " by downloading the latest package from: https://github.com/hornbill/" + appName + "/releases/tag/v" + res.Current
		logger(5, msg, true, true)
	}
}

// espLogger -- Log to ESP
func espLogger(message string, severity string) {
	if configDryrun {
		message = "[DRYRUN] " + message
	}
	espXmlmc.SetParam("fileName", appName)
	espXmlmc.SetParam("group", "general")
	espXmlmc.SetParam("severity", severity)
	espXmlmc.SetParam("message", message)
	espXmlmc.Invoke("system", "logMessage")
}

func logger(t int, s string, outputToCLI, outputToESP bool) {
	//-- Create Log Entry
	var espLogType string
	switch t {
	case 1:
		espLogType = "debug"
	case 4:
		espLogType = "error"
	case 5:
		espLogType = "warn"
	default:
		espLogType = "notice"
	}
	if outputToESP {
		espLogger(s, espLogType)
	}
	hornbillHelpers.Logger(t, s, outputToCLI, logFileName)
}
