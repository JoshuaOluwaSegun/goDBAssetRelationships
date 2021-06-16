package main

import (
	"encoding/xml"
	"errors"
	"fmt"

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
		logger(1, "No existing asset impacts could be found", true, true)
		return nil
	}
	var i int
	logger(1, "Retrieving "+fmt.Sprint(assetImpactCount)+" asset impacts from Hornbill. Please wait...", true, true)

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
	logger(1, fmt.Sprint(len(assetImpacts))+" asset impact records cached.", true, true)
	return err
}

func getAssetImpactCount() (int, error) {
	espXmlmc.SetParam("application", "com.hornbill.servicemanager")
	espXmlmc.SetParam("table", "h_cmdb_config_items_impact")
	espXmlmc.SetParam("where", "h_entity_l_name = 'asset' AND h_entity_r_name = 'asset'")
	if configDryrun {
		logger(3, "[DRYRUN] [IMPACT] [COUNT] "+espXmlmc.GetParam(), false, false)
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
	espXmlmc.SetParam("application", "com.hornbill.servicemanager")
	espXmlmc.SetParam("queryName", "getImpactsForExplorer")
	espXmlmc.OpenElement("queryParams")
	espXmlmc.SetParam("rowstart", fmt.Sprint(rowStart))
	espXmlmc.SetParam("limit", fmt.Sprint(limit))
	espXmlmc.CloseElement("queryParams")
	if configDryrun {
		logger(3, "[DRYRUN] [IMPACT] [GET] "+espXmlmc.GetParam(), false, false)
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
	espXmlmc.SetParam("application", "com.hornbill.servicemanager")
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
		logger(3, "[DRYRUN] [IMPACT] [CREATE] "+espXmlmc.GetParam(), false, false)
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
	espXmlmc.SetParam("application", "com.hornbill.servicemanager")
	espXmlmc.SetParam("entity", "ConfigurationItemsImpact")
	espXmlmc.OpenElement("primaryEntityData")
	espXmlmc.OpenElement("record")
	espXmlmc.SetParam("h_pk_confitemimpactid", id)
	espXmlmc.SetParam("h_impact", impact)
	espXmlmc.CloseElement("record")
	espXmlmc.CloseElement("primaryEntityData")
	if configDryrun {
		logger(3, "[DRYRUN] [IMPACT] [UPDATE] "+espXmlmc.GetParam(), false, false)
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

func deleteImpact(id string) error {
	espXmlmc.SetParam("application", "com.hornbill.servicemanager")
	espXmlmc.SetParam("entity", "ConfigurationItemsImpact")
	espXmlmc.SetParam("keyValue", id)
	if configDryrun {
		logger(3, "[DRYRUN] [IMPACT] [DELETE] "+espXmlmc.GetParam(), false, false)
		espXmlmc.ClearParam()
		return nil
	}
	linkAssetResult, err := espXmlmc.Invoke("data", "entityDeleteRecord")
	if err != nil {
		retError := "deleteImpact:Invoke:" + err.Error()
		return errors.New(retError)
	}

	var xmlResponse methodCallResult
	err = xml.Unmarshal([]byte(linkAssetResult), &xmlResponse)
	if err != nil {
		retError := "deleteImpact:Unmarshal:" + err.Error()
		return errors.New(retError)
	}
	if xmlResponse.Status != "ok" {
		retError := "deleteImpact:Invoke:" + xmlResponse.State.ErrorRet
		return errors.New(retError)
	}
	return nil
}
