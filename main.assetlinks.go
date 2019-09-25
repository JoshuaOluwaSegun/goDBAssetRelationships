package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strings"

	hornbillHelpers "github.com/hornbill/goHornbillHelpers"
	"github.com/hornbill/pb"
)

//cacheAssetLinks  - caches asset records from instance
func cacheAssetLinks() error {
	//Get Count
	var err error
	assetLinkCount, err := getAssetLinkCount()
	if err != nil {
		return err
	}

	if assetLinkCount == 0 {
		hornbillHelpers.Logger(1, "No existing asset links could be found", true, logFileName)
		return nil
	}
	var i int
	hornbillHelpers.Logger(1, "Retrieving "+fmt.Sprint(assetLinkCount)+" asset entity links from Hornbill. Please wait...", true, logFileName)

	bar := pb.New(assetLinkCount)
	bar.ShowPercent = false
	bar.ShowCounters = false
	bar.ShowTimeLeft = false
	bar.Start()
	assetPrefix := "urn:sys:entity:com.hornbill.servicemanager:Asset:"
	for i = 0; i <= assetLinkCount; i += xmlmcPageSize {
		blockAssetLinks, err := getAssetLinks(i, xmlmcPageSize)
		if err != nil {
			bar.Finish()
			return err
		}
		if len(blockAssetLinks) > 0 {
			for _, v := range blockAssetLinks {
				if strings.HasPrefix(v.IDL, assetPrefix) && strings.HasPrefix(v.IDL, assetPrefix) {
					concatedAssets := strings.Replace(v.IDL, assetPrefix, "", 1) + ":" + strings.Replace(v.IDR, assetPrefix, "", 1)
					assetLinks[concatedAssets] = v
				}
			}
		}
		bar.Add(xmlmcPageSize)
	}
	bar.Finish()
	hornbillHelpers.Logger(1, fmt.Sprint(len(assetLinks))+" asset links cached.", true, logFileName)
	return err
}

func getAssetLinkCount() (int, error) {
	espXmlmc.SetParam("application", "com.hornbill.servicemanager")
	espXmlmc.SetParam("table", "h_cmdb_links")
	espXmlmc.SetParam("where", "h_rel_type_l = 1 AND h_rel_type_r = 1")
	if configDryrun {
		hornbillHelpers.Logger(3, "[DRYRUN] [LINK] [COUNT] "+espXmlmc.GetParam(), false, logFileName)
	}
	xmlAssetLinksCount, err := espXmlmc.Invoke("data", "getRecordCount")
	if err != nil {
		retError := "getAssetLinkCount:Invoke:" + err.Error()
		return 0, errors.New(retError)
	}

	var xmlResponse methodCallResult
	err = xml.Unmarshal([]byte(xmlAssetLinksCount), &xmlResponse)
	if err != nil {
		retError := "getAssetLinkCount:Unmarshal:" + err.Error()
		return 0, errors.New(retError)
	}
	if xmlResponse.Status != "ok" {
		retError := "getAssetLinkCount:Xmlmc:" + xmlResponse.State.ErrorRet
		return 0, errors.New(retError)
	}
	return xmlResponse.Params.Count, err
}

func getAssetLinks(rowStart, limit int) ([]assetLinkStruct, error) {
	var assetLinksBlock []assetLinkStruct
	espXmlmc.SetParam("application", "com.hornbill.servicemanager")
	espXmlmc.SetParam("queryName", "assetLinks")
	espXmlmc.OpenElement("queryParams")
	espXmlmc.SetParam("rowstart", fmt.Sprint(rowStart))
	espXmlmc.SetParam("limit", fmt.Sprint(limit))
	espXmlmc.CloseElement("queryParams")
	if configDryrun {
		hornbillHelpers.Logger(3, "[DRYRUN] [LINK] [GET] "+espXmlmc.GetParam(), false, logFileName)
	}
	xmlAssets, err := espXmlmc.Invoke("data", "queryExec")
	if err != nil {
		retError := "getAssetLinks:Invoke:" + err.Error()
		return assetLinksBlock, errors.New(retError)
	}

	var xmlResponse methodCallResultLinks
	err = xml.Unmarshal([]byte(xmlAssets), &xmlResponse)
	if err != nil {
		retError := "getAssetLinks:Unmarshal:" + err.Error()
		return assetLinksBlock, errors.New(retError)
	}
	if xmlResponse.Status != "ok" {
		retError := "getAssetLinks:Xmlmc:" + xmlResponse.State.ErrorRet
		return assetLinksBlock, errors.New(retError)
	}
	return xmlResponse.Links, err
}

func linkAsset(lid, rid string) error {
	espXmlmc.SetParam("leftEntityId", lid)
	espXmlmc.SetParam("leftEntityType", "Asset")
	espXmlmc.SetParam("leftRelType", "1")
	espXmlmc.SetParam("rightEntityId", rid)
	espXmlmc.SetParam("rightEntityType", "Asset")
	espXmlmc.SetParam("rightRelType", "1")
	espXmlmc.SetParam("dependsOn", "0")
	if configDryrun {
		hornbillHelpers.Logger(3, "[DRYRUN] [LINK] [CREATE] "+espXmlmc.GetParam(), false, logFileName)
		espXmlmc.ClearParam()
		return nil
	}

	linkAssetResult, err := espXmlmc.Invoke("apps/com.hornbill.servicemanager/Asset", "linkAsset")

	if err != nil {
		retError := "linkAsset:Invoke:" + err.Error()
		return errors.New(retError)
	}

	var xmlResponse methodCallResult
	err = xml.Unmarshal([]byte(linkAssetResult), &xmlResponse)
	if err != nil {
		retError := "linkAsset:Unmarshal:" + err.Error()
		return errors.New(retError)
	}
	if xmlResponse.Status != "ok" {
		retError := "linkAsset:Xmlmc:" + xmlResponse.State.ErrorRet
		return errors.New(retError)
	}
	return nil
}
