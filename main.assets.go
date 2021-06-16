package main

import (
	"encoding/xml"
	"errors"
	"fmt"

	"github.com/hornbill/pb"
)

//cacheAssets  - caches asset records from instance
func cacheAssets() error {
	//Get Count
	var err error
	assetCount, err = getAssetCount()
	if err != nil {
		return err
	}

	if assetCount == 0 {
		return errors.New("no assets could be found on your hornbill instance")
	}
	var i int
	logger(1, "Retrieving "+fmt.Sprint(assetCount)+" assets from Hornbill. Please wait...", true, true)

	bar := pb.New(assetCount)
	bar.ShowPercent = false
	bar.ShowCounters = false
	bar.ShowTimeLeft = false
	bar.Start()
	for i = 0; i <= assetCount; i += xmlmcPageSize {
		blockAssets, err := getAssets(i, xmlmcPageSize)
		if err != nil {
			bar.Finish()
			return err
		}
		if len(blockAssets) > 0 {
			for _, v := range blockAssets {
				keyval := getKeyVal(&v)
				assets[keyval] = v
			}
		}
		bar.Add(xmlmcPageSize)
	}
	bar.Finish()
	logger(1, fmt.Sprint(len(assets))+" assets cached.", true, true)
	return err
}

func getKeyVal(asset *assetDetailsStruct) string {
	switch importConf.AssetIdentifier.Hornbill {
	case "PrimaryKey":
		return asset.AssetID
	case "Description":
		return asset.AssetDescription
	case "Name":
		return asset.AssetName
	case "Tag":
		return asset.AssetTag
	}
	return asset.AssetName
}

func getAssetCount() (int, error) {
	espXmlmc.SetParam("application", "com.hornbill.servicemanager")
	espXmlmc.SetParam("table", "h_cmdb_assets")

	if configDryrun {
		logger(3, "[DRYRUN] [ASSETS] [COUNT] "+espXmlmc.GetParam(), false, false)
	}
	xmlAssetCount, err := espXmlmc.Invoke("data", "getRecordCount")
	if err != nil {
		retError := "getAssetCount:Invoke:" + err.Error()
		return 0, errors.New(retError)
	}

	var xmlResponse methodCallResult
	err = xml.Unmarshal([]byte(xmlAssetCount), &xmlResponse)
	if err != nil {
		retError := "getAssetCount:Unmarshal:" + err.Error()
		return 0, errors.New(retError)
	}
	if xmlResponse.Status != "ok" {
		retError := "getAssetCount:Xmlmc:" + xmlResponse.State.ErrorRet
		return 0, errors.New(retError)
	}
	return xmlResponse.Params.Count, err
}

func getAssets(rowStart, limit int) ([]assetDetailsStruct, error) {
	var assets []assetDetailsStruct
	espXmlmc.SetParam("application", "com.hornbill.servicemanager")
	espXmlmc.SetParam("queryName", "getAssetsList")
	espXmlmc.OpenElement("queryParams")
	espXmlmc.SetParam("rowstart", fmt.Sprint(rowStart))
	espXmlmc.SetParam("limit", fmt.Sprint(limit))
	espXmlmc.CloseElement("queryParams")
	if configDryrun {
		logger(3, "[DRYRUN] [ASSETS] [GET] "+espXmlmc.GetParam(), false, false)
	}
	xmlAssets, err := espXmlmc.Invoke("data", "queryExec")
	if err != nil {
		retError := "getAssets:Invoke:" + err.Error()
		return assets, errors.New(retError)
	}

	var xmlResponse methodCallResult
	err = xml.Unmarshal([]byte(xmlAssets), &xmlResponse)
	if err != nil {
		retError := "getAssets:Unmarshal:" + err.Error()
		return assets, errors.New(retError)
	}
	if xmlResponse.Status != "ok" {
		retError := "getAssets:Xmlmc:" + xmlResponse.State.ErrorRet
		return assets, errors.New(retError)
	}
	return xmlResponse.Params.Assets, err
}
