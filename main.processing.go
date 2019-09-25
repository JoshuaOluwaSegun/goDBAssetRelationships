package main

import (
	"fmt"
	"strconv"

	hornbillHelpers "github.com/hornbill/goHornbillHelpers"
	"github.com/hornbill/pb"
)

func processRelationships() {

	hornbillHelpers.Logger(1, "Processing "+strconv.Itoa(len(assetRelationships))+" found relationship records...", true, logFileName)
	bar := pb.New(len(assetRelationships))
	bar.ShowPercent = false
	bar.ShowCounters = true
	bar.ShowTimeLeft = false
	bar.Start()

	for _, rel := range assetRelationships {
		bar.Increment()
		parentName := fmt.Sprintf("%s", rel[importConf.AssetIdentifier.Parent])
		childName := fmt.Sprintf("%s", rel[importConf.AssetIdentifier.Child])
		parentAssetID := getAssetID(parentName)
		childAssetID := getAssetID(childName)
		if parentAssetID == "" {
			hornbillHelpers.Logger(5, "Could not find Parent asset: ["+parentName+"]", false, logFileName)
			continue
		}
		if parentAssetID == "" {
			hornbillHelpers.Logger(5, "Could not find Child asset: ["+childName+"]", false, logFileName)
			continue
		}

		hornbillHelpers.Logger(1, "Processing "+parentName+" ["+parentAssetID+"] to "+childAssetID+" ["+childName+"]", false, logFileName)

		//Process Service Manager asset link first
		pcLinkIDs := parentAssetID + ":" + childAssetID
		cpLinkIDs := childAssetID + ":" + parentAssetID
		_, pcok := assetLinks[pcLinkIDs]
		_, cpok := assetLinks[cpLinkIDs]

		if !cpok && !pcok {
			//Link doesn't exist, go add it
			err := linkAsset(parentAssetID, childAssetID)
			if err != nil {
				counters.linksFailed++
				hornbillHelpers.Logger(4, err.Error(), false, logFileName)
				continue
			} else {
				counters.linksCreated++
				if !configDryrun {
					hornbillHelpers.Logger(1, "Linked successfully", false, logFileName)
				}
			}
		} else {
			counters.linksSkipped++
			hornbillHelpers.Logger(1, "Link already exists between assets", false, logFileName)
		}

		if !configManager {
			//Break out if config manager isn't installed
			continue
		}

		//Sort out dependency record
		recDependency := fmt.Sprintf("%s", rel[importConf.AssetIdentifier.Dependency])
		dependency, depMapped := importConf.DepencencyMapping[recDependency]
		if !depMapped {
			hornbillHelpers.Logger(5, "Dependency ["+recDependency+"] not found in mapping, so using ["+recDependency+"]", false, logFileName)
			dependency = recDependency
		}
		depRecord, pcdepok := assetDependencies[pcLinkIDs]
		if !pcdepok {
			//Dependency doesn't exist - add it
			err := addDependency(parentAssetID, childAssetID, dependency)
			if err != nil {
				counters.depsFailed++
				hornbillHelpers.Logger(4, err.Error(), false, logFileName)
				continue
			} else {
				counters.depsCreated++
				if !configDryrun {
					hornbillHelpers.Logger(1, "Dependency ["+dependency+"] created sucessfully", false, logFileName)
				}
			}
		} else {
			//Check dependency for match
			if depRecord.Dependency != dependency {
				err := updateDependency(depRecord.ID, dependency)
				if err != nil {
					counters.depsUpdateFailed++
					hornbillHelpers.Logger(4, err.Error(), false, logFileName)
				} else {
					counters.depsUpdated++
					if !configDryrun {
						hornbillHelpers.Logger(1, "Dependency ["+dependency+"] updated successfully", false, logFileName)
					}
				}

			} else {
				counters.depsSkipped++
				hornbillHelpers.Logger(1, "Dependency ["+dependency+"] already exists between assets", false, logFileName)
			}
		}

		//Sort out impact record
		recImpact := fmt.Sprintf("%s", rel[importConf.AssetIdentifier.Impact])
		impact, impMapped := importConf.ImpactMapping[recImpact]
		if !impMapped {
			hornbillHelpers.Logger(5, "Impact ["+recImpact+"] not found in mapping.", false, logFileName)
			impact = recImpact
		}
		impRecord, pcimpok := assetImpacts[pcLinkIDs]
		if !pcimpok {
			//Impact doesn't exist - add it
			err := addImpact(parentAssetID, childAssetID, impact)
			if err != nil {
				counters.impsFailed++
				hornbillHelpers.Logger(4, err.Error(), false, logFileName)
				continue
			} else {
				counters.impsCreated++
				if !configDryrun {
					hornbillHelpers.Logger(1, "Impact ["+impact+"] created successfully", false, logFileName)
				}
			}
		} else {
			//Check impact for match
			if impRecord.Impact != impact {
				err := updateImpact(impRecord.ID, impact)
				if err != nil {
					counters.impsUpdateFailed++
					hornbillHelpers.Logger(4, err.Error(), false, logFileName)
				} else {
					counters.impsUpdated++
					if !configDryrun {
						hornbillHelpers.Logger(1, "Impact ["+impact+"] updated successfully", false, logFileName)
					}
				}
			} else {
				counters.impsSkipped++
				hornbillHelpers.Logger(1, "Impact ["+impact+"] already exists between assets", false, logFileName)
			}
		}

	}
	bar.Finish()
}

//getAssetID -- Check if asset exists
func getAssetID(assetIdentifier string) string {
	assetRecord, ok := assets[assetIdentifier]
	if ok {
		return assetRecord.AssetID
	}
	return ""
}
