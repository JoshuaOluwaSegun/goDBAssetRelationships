package main

import (
	"encoding/xml"
	"errors"
	"fmt"

	hornbillHelpers "github.com/hornbill/goHornbillHelpers"
	"github.com/hornbill/pb"
)

//cacheAssetImpacts  - caches asset impact records from instance
func cacheAssetImpacts() error {
	//Get Count
	var err error
	assetImpactCount, err := getAssetImpactCount()
	if err != nil {
		return err
	}

	if assetImpactCount == 0 {
		hornbillHelpers.Logger(1, "No existing asset impacts could be found", true, logFileName)
		return nil
	}
	var i int
	hornbillHelpers.Logger(1, "Retrieving "+fmt.Sprint(assetImpactCount)+" asset impacts from Hornbill. Please wait...", true, logFileName)

	bar := pb.New(assetImpactCount)
	bar.ShowPercent = false
	bar.ShowCounters = false
	bar.ShowTimeLeft = false
	bar.Start()
	for i = 0; i <= assetImpactCount; i += xmlmcPageSize {
		blockAssetDeps, err := getAssetImpacts(i, xmlmcPageSize)
		if err != nil {
			bar.Finish()
			return err
		}
		if len(blockAssetDeps) > 0 {
			for _, v := range blockAssetDeps {
				concatedAssets := v.LID + ":" + v.RID
				assetImpacts[concatedAssets] = v
			}
		}
		bar.Add(xmlmcPageSize)
	}
	bar.Finish()
	hornbillHelpers.Logger(1, fmt.Sprint(len(assetImpacts))+" asset impact records cached.", true, logFileName)
	return err
}

func getAssetImpactCount() (int, error) {
	espXmlmc.SetParam("application", "com.hornbill.configurationmanager")
	espXmlmc.SetParam("table", "h_cmdb_config_items_impact")
	espXmlmc.SetParam("where", "h_entity_l_name = 'asset' AND h_entity_r_name = 'asset'")
	if configDryrun {
		hornbillHelpers.Logger(3, "[DRYRUN] [IMPACT] [COUNT] "+espXmlmc.GetParam(), false, logFileName)
	}
	xmlAssetLinksCount, err := espXmlmc.Invoke("data", "getRecordCount")
	if err != nil {
		retError := "getAssetImpactCount:Invoke:" + err.Error()
		return 0, errors.New(retError)
	}

	var xmlResponse methodCallResult
	err = xml.Unmarshal([]byte(xmlAssetLinksCount), &xmlResponse)
	if err != nil {
		retError := "getAssetImpactCount:Unmarshal:" + err.Error()
		return 0, errors.New(retError)
	}
	if xmlResponse.Status != "ok" {
		retError := "getAssetImpactCount:Xmlmc:" + xmlResponse.State.ErrorRet
		return 0, errors.New(retError)
	}
	return xmlResponse.Params.Count, err
}

func getAssetImpacts(rowStart, limit int) ([]assetImpactStruct, error) {
	var assetImpactsBlock []assetImpactStruct
	espXmlmc.SetParam("application", "com.hornbill.configurationmanager")
	espXmlmc.SetParam("queryName", "getImpacts")
	espXmlmc.OpenElement("queryParams")
	espXmlmc.SetParam("rowstart", fmt.Sprint(rowStart))
	espXmlmc.SetParam("limit", fmt.Sprint(limit))
	espXmlmc.CloseElement("queryParams")
	if configDryrun {
		hornbillHelpers.Logger(3, "[DRYRUN] [IMPACT] [GET] "+espXmlmc.GetParam(), false, logFileName)
	}
	xmlAssets, err := espXmlmc.Invoke("data", "queryExec")
	if err != nil {
		retError := "getAssetImpacts:Invoke:" + err.Error()
		return assetImpactsBlock, errors.New(retError)
	}

	var xmlResponse methodCallResultImpacts
	err = xml.Unmarshal([]byte(xmlAssets), &xmlResponse)
	if err != nil {
		retError := "getAssetImpacts:Unmarshal:" + err.Error()
		return assetImpactsBlock, errors.New(retError)
	}
	if xmlResponse.Status != "ok" {
		retError := "getAssetImpacts:Xmlmc:" + xmlResponse.State.ErrorRet
		return assetImpactsBlock, errors.New(retError)
	}
	return xmlResponse.Impacts, err
}

func addImpact(lid, rid, impact string) error {
	espXmlmc.SetParam("application", "com.hornbill.configurationmanager")
	espXmlmc.SetParam("entity", "ConfigurationItemsImpact")
	espXmlmc.OpenElement("primaryEntityData")
	espXmlmc.OpenElement("record")
	espXmlmc.SetParam("h_entity_l_id", lid)
	espXmlmc.SetParam("h_entity_l_name", "asset")
	espXmlmc.SetParam("h_entity_r_id", rid)
	espXmlmc.SetParam("h_entity_r_name", "asset")
	espXmlmc.SetParam("h_impact", impact)
	espXmlmc.CloseElement("record")
	espXmlmc.CloseElement("primaryEntityData")
	if configDryrun {
		hornbillHelpers.Logger(3, "[DRYRUN] [IMPACT] [CREATE] "+espXmlmc.GetParam(), false, logFileName)
		espXmlmc.ClearParam()
		return nil
	}
	linkAssetResult, err := espXmlmc.Invoke("data", "entityAddRecord")
	if err != nil {
		retError := "addImpact:Invoke:" + err.Error()
		return errors.New(retError)
	}

	var xmlResponse methodCallResult
	err = xml.Unmarshal([]byte(linkAssetResult), &xmlResponse)
	if err != nil {
		retError := "addImpact:Unmarshal:" + err.Error()
		return errors.New(retError)
	}
	if xmlResponse.Status != "ok" {
		retError := "addImpact:Invoke:" + xmlResponse.State.ErrorRet
		return errors.New(retError)
	}
	return nil
}

func updateImpact(id, impact string) error {
	espXmlmc.SetParam("application", "com.hornbill.configurationmanager")
	espXmlmc.SetParam("entity", "ConfigurationItemsImpact")
	espXmlmc.OpenElement("primaryEntityData")
	espXmlmc.OpenElement("record")
	espXmlmc.SetParam("h_pk_confitemimpactid", id)
	espXmlmc.SetParam("h_impact", impact)
	espXmlmc.CloseElement("record")
	espXmlmc.CloseElement("primaryEntityData")
	if configDryrun {
		hornbillHelpers.Logger(3, "[DRYRUN] [IMPACT] [UPDATE] "+espXmlmc.GetParam(), false, logFileName)
		espXmlmc.ClearParam()
		return nil
	}
	linkAssetResult, err := espXmlmc.Invoke("data", "entityUpdateRecord")
	if err != nil {
		retError := "updateImpact:Invoke:" + err.Error()
		return errors.New(retError)
	}

	var xmlResponse methodCallResult
	err = xml.Unmarshal([]byte(linkAssetResult), &xmlResponse)
	if err != nil {
		retError := "updateImpact:Unmarshal:" + err.Error()
		return errors.New(retError)
	}
	if xmlResponse.Status != "ok" {
		retError := "updateImpact:Invoke:" + xmlResponse.State.ErrorRet
		return errors.New(retError)
	}
	return nil
}
