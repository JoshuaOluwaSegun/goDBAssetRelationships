package main

import (
	"fmt"
	"strconv"

	"github.com/hornbill/pb"
)

func processRelationships() {

	logger(1, "Processing "+strconv.Itoa(len(assetRelationships))+" found relationship records...", true, true)
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
			logger(5, "Could not find Parent asset: ["+parentName+"]", false, false)
			continue
		}
		if childAssetID == "" {
			logger(5, "Could not find Child asset: ["+childName+"]", false, false)
			continue
		}

		logger(1, "Processing "+parentName+" ["+parentAssetID+"] to "+childAssetID+" ["+childName+"]", false, false)

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
				logger(4, err.Error(), false, true)
				continue
			} else {
				counters.linksCreated++
				if !configDryrun {
					logger(1, "Linked successfully", false, false)
				}
			}
		} else {
			counters.linksSkipped++
			logger(1, "Link already exists between assets", false, false)
		}

		//Sort out dependency record
		recDependency := fmt.Sprintf("%s", rel[importConf.AssetIdentifier.Dependency])
		dependency, depMapped := importConf.DepencencyMapping[recDependency]
		if !depMapped {
			logger(5, "Dependency ["+recDependency+"] not found in mapping, so using ["+recDependency+"]", false, false)
			dependency = recDependency
		}
		depRecord, pcdepok := assetDependencies[pcLinkIDs]
		if !pcdepok {
			//Dependency doesn't exist - add it
			err := addDependency(parentAssetID, childAssetID, dependency)
			if err != nil {
				counters.depsFailed++
				logger(4, err.Error(), false, true)
				continue
			} else {
				counters.depsCreated++
				if !configDryrun {
					logger(1, "Dependency ["+dependency+"] created sucessfully", false, false)
				}
			}
		} else {
			//Check dependency for match
			if depRecord.Dependency != dependency {
				err := updateDependency(depRecord.ID, dependency)
				if err != nil {
					counters.depsUpdateFailed++
					logger(4, err.Error(), false, true)
				} else {
					counters.depsUpdated++
					if !configDryrun {
						logger(1, "Dependency ["+dependency+"] updated successfully", false, false)
					}
				}

			} else {
				counters.depsSkipped++
				logger(1, "Dependency ["+dependency+"] already exists between assets", false, false)
			}
		}

		//Sort out impact record
		recImpact := fmt.Sprintf("%s", rel[importConf.AssetIdentifier.Impact])
		impact, impMapped := importConf.ImpactMapping[recImpact]
		if !impMapped {
			logger(5, "Impact ["+recImpact+"] not found in mapping.", false, false)
			impact = recImpact
		}
		impRecord, pcimpok := assetImpacts[pcLinkIDs]
		if !pcimpok {
			//Impact doesn't exist - add it
			err := addImpact(parentAssetID, childAssetID, impact)
			if err != nil {
				counters.impsFailed++
				logger(4, err.Error(), false, true)
				continue
			} else {
				counters.impsCreated++
				if !configDryrun {
					logger(1, "Impact ["+impact+"] created successfully", false, false)
				}
			}
		} else {
			//Check impact for match
			if impRecord.Impact != impact {
				err := updateImpact(impRecord.ID, impact)
				if err != nil {
					counters.impsUpdateFailed++
					logger(4, err.Error(), false, true)
				} else {
					counters.impsUpdated++
					if !configDryrun {
						logger(1, "Impact ["+impact+"] updated successfully", false, false)
					}
				}
			} else {
				counters.impsSkipped++
				logger(1, "Impact ["+impact+"] already exists between assets", false, false)
			}
		}

	}
	bar.Finish()
}

func processRelationshipRemovals() {

	logger(1, "Processing "+strconv.Itoa(len(assetDeleteRelationships))+" found relationship removal records...", true, true)
	bar := pb.New(len(assetDeleteRelationships))
	bar.ShowPercent = false
	bar.ShowCounters = true
	bar.ShowTimeLeft = false
	bar.Start()

	for _, rel := range assetDeleteRelationships {
		bar.Increment()
		parentName := fmt.Sprintf("%s", rel[importConf.RemoveAssetIdentifier.Parent])
		childName := fmt.Sprintf("%s", rel[importConf.RemoveAssetIdentifier.Child])
		parentAssetID := getAssetID(parentName)
		childAssetID := getAssetID(childName)
		if parentAssetID == "" {
			logger(5, "Could not find Parent asset: ["+parentName+"]", false, false)
			continue
		}
		if childAssetID == "" {
			logger(5, "Could not find Child asset: ["+childName+"]", false, false)
			continue
		}

		logger(1, "Processing removal of "+parentName+" ["+parentAssetID+"] link to "+childAssetID+" ["+childName+"]", false, false)

		//Process Service Manager asset link first
		pcLinkIDs := parentAssetID + ":" + childAssetID
		cpLinkIDs := childAssetID + ":" + parentAssetID
		_, pcok := assetLinks[pcLinkIDs]
		_, cpok := assetLinks[cpLinkIDs]

		if !cpok && !pcok {
			counters.removeLinksSkipped++
			logger(1, "Link doesn't exist between assets", false, false)
		} else {
			//Link doesn't exist, go add it
			err := unlinkAsset(parentAssetID, childAssetID)
			if err != nil {
				counters.removeLinksFailed++
				logger(4, err.Error(), false, true)
				continue
			} else {
				counters.removeLinksSuccess++
				if !configDryrun {
					logger(1, "Unlinked successfully", false, false)
				}
			}
		}

		//Sort out dependency record
		recDependency := fmt.Sprintf("%s", rel[importConf.RemoveAssetIdentifier.Dependency])
		dependency, depMapped := importConf.DepencencyMapping[recDependency]
		if !depMapped {
			logger(5, "Dependency ["+recDependency+"] not found in mapping, so using ["+recDependency+"]", false, false)
			dependency = recDependency
		}
		depRecord, pcdepok := assetDependencies[pcLinkIDs]
		if !pcdepok {
			//Dependency doesn't exist
			logger(1, "Dependency ["+dependency+"] doesn't exist", false, false)
			counters.removeDepsSkipped++
		} else {
			//Check dependency for match
			if depRecord.Dependency == dependency {
				err := deleteDependency(depRecord.ID)
				if err != nil {
					counters.removeDepsFailed++
					logger(4, err.Error(), false, true)
				} else {
					counters.removeDepsSuccess++
					if !configDryrun {
						logger(1, "Dependency ["+dependency+"] removed successfully", false, false)
					}
				}

			} else {
				counters.removeDepsSkipped++
				logger(1, "Dependency ["+dependency+"] doesn't match record dependency type ["+depRecord.Dependency+"]", false, false)
			}
		}

		//Sort out impact record
		recImpact := fmt.Sprintf("%s", rel[importConf.RemoveAssetIdentifier.Impact])
		impact, impMapped := importConf.ImpactMapping[recImpact]
		if !impMapped {
			logger(5, "Impact ["+recImpact+"] not found in mapping.", false, false)
			impact = recImpact
		}
		impRecord, pcimpok := assetImpacts[pcLinkIDs]
		if !pcimpok {
			//Impact doesn't exist
			logger(1, "Impact ["+impact+"] doesn't exist", false, false)
			counters.removeImpsSkipped++
		} else {
			//Check impact for match
			if impRecord.Impact == impact {
				err := deleteImpact(impRecord.ID)
				if err != nil {
					counters.removeImpsFailed++
					logger(4, err.Error(), false, true)
				} else {
					counters.removeImpsSuccess++
					if !configDryrun {
						logger(1, "Impact ["+impact+"] removed successfully", false, false)
					}
				}
			} else {
				counters.removeImpsSkipped++
				logger(1, "Impact ["+impact+"] doesn't match record impact type ["+impRecord.Impact+"]", false, false)
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
