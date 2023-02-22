package main

import (
	apiLib "github.com/hornbill/goApiLib"
)

// ----- Constants -----
const (
	version       = "1.3.0"
	xmlmcPageSize = 100
	appName       = "goDBAssetRelationships"
)

// ----- Variables -----
var (
	assetCount               int
	assets                   = make(map[string]assetDetailsStruct)
	assetLinks               = make(map[string]assetLinkStruct)
	assetDependencies        = make(map[string]assetDependencyStruct)
	assetImpacts             = make(map[string]assetImpactStruct)
	assetRelationships       []map[string]interface{}
	assetDeleteRelationships []map[string]interface{}
	counters                 counterTypeStruct
	configDryrun             bool
	configFileName           string
	configVersion            bool
	espXmlmc                 *apiLib.XmlmcInstStruct
	importConf               sqlImportConfStruct
	logFileName              string
	timeNow                  string
)

type counterTypeStruct struct {
	linksCreated       int
	linksSkipped       int
	linksFailed        int
	depsCreated        int
	depsUpdated        int
	depsSkipped        int
	depsUpdateFailed   int
	depsFailed         int
	impsCreated        int
	impsUpdated        int
	impsSkipped        int
	impsUpdateFailed   int
	impsFailed         int
	removeLinksSuccess int
	removeLinksSkipped int
	removeLinksFailed  int
	removeDepsSuccess  int
	removeDepsSkipped  int
	removeDepsFailed   int
	removeImpsSuccess  int
	removeImpsSkipped  int
	removeImpsFailed   int
}

// -- Config Structs
type sqlImportConfStruct struct {
	APIKey                string
	InstanceID            string
	LogSizeBytes          int64
	DBConf                sqlConfStruct
	Query                 string
	AssetIdentifier       assetIdentifierStruct
	DepencencyMapping     map[string]string
	ImpactMapping         map[string]string
	RemoveLinks           bool
	RemoveQuery           string
	RemoveAssetIdentifier assetIdentifierStruct
}

type sqlConfStruct struct {
	Driver         string
	Server         string
	Database       string
	Authentication string
	UserName       string
	Password       string
	Port           int
	Encrypt        bool
}

type assetIdentifierStruct struct {
	Parent          string
	Child           string
	Dependency      string
	Impact          string
	Hornbill        string
	RemoveBothSides bool
}

// -- XMLMC Call Structs
type methodCallResult struct {
	State  stateStruct  `xml:"state"`
	Status string       `xml:"status,attr"`
	Params paramsStruct `xml:"params"`
}
type stateStruct struct {
	Code     string `xml:"code"`
	ErrorRet string `xml:"error"`
}
type paramsStruct struct {
	Count  int                  `xml:"count"`
	Assets []assetDetailsStruct `xml:"rowData>row"`
}

type assetDetailsStruct struct {
	AssetID          string `xml:"h_pk_asset_id"`
	AssetDescription string `xml:"asset_description"`
	AssetName        string `xml:"asset_name"`
	AssetTag         string `xml:"h_asset_tag"`
}

type methodCallResultLinks struct {
	State  stateStruct       `xml:"state"`
	Status string            `xml:"status,attr"`
	Links  []assetLinkStruct `xml:"params>rowData>row"`
}
type assetLinkStruct struct {
	ID       string `xml:"h_pk_id"`
	IDL      string `xml:"h_fk_id_l"`
	IDR      string `xml:"h_fk_id_r"`
	RelTypeL string `xml:"h_rel_type_l"`
	RelTypeR string `xml:"h_rel_type_r"`
	OpDep    string `xml:"h_op_dep"`
}

type methodCallResultDependencies struct {
	State        stateStruct             `xml:"state"`
	Status       string                  `xml:"status,attr"`
	Dependencies []assetDependencyStruct `xml:"params>rowData>row"`
}
type assetDependencyStruct struct {
	ID         string `xml:"h_pk_confitemdependencyid"`
	LID        string `xml:"h_entity_l_id"`
	LName      string `xml:"h_entity_l_name"`
	RID        string `xml:"h_entity_r_id"`
	RName      string `xml:"h_entity_r_name"`
	Dependency string `xml:"h_dependency"`
}

type methodCallResultImpacts struct {
	State   stateStruct         `xml:"state"`
	Status  string              `xml:"status,attr"`
	Impacts []assetImpactStruct `xml:"params>rowData>row"`
}
type assetImpactStruct struct {
	ID     string `xml:"h_pk_confitemimpactid"`
	LID    string `xml:"h_entity_l_id"`
	LName  string `xml:"h_entity_l_name"`
	RID    string `xml:"h_entity_r_id"`
	RName  string `xml:"h_entity_r_name"`
	Impact string `xml:"h_impact"`
}
