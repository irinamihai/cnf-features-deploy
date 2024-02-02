package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	siteConfigs "github.com/openshift-kni/cnf-features-deploy/ztp/siteconfig-generator/siteConfig"
	"gopkg.in/yaml.v3"
)

func main() {
	localExtraManifestPath := flag.String("manifestPath", "", "Directory with pre-defined extra manifest")
	extraManifestOnly := flag.Bool("extraManifestOnly", false, "Generate extra manifests only")
	outPath := flag.String("outPath", siteConfigs.UnsetStringValue, "Directory to write the generated installation resources")
	// Parse command input
	flag.Parse()

	// Collect and parse siteconfig files paths
	siteConfigFiles := flag.Args()
	scBuilder, _ := siteConfigs.NewSiteConfigBuilder()
	if *localExtraManifestPath != "" {
		scBuilder.SetLocalExtraManifestPath(*localExtraManifestPath)
	}

	// Set up the resources to hold errors.	scBuilder.BuildErrorCRs()
	errorNs, errorConfigMap, err:= scBuilder.BuildErrorCRs()
	if err != nil {
		log.Fatalf("Error: could not create error resources %s\n", err)
	}

	for _, siteConfigFile := range siteConfigFiles {
		fileData, err := siteConfigs.ReadFile(siteConfigFile)
		if err != nil {
			errorMsg := fmt.Sprintf("Error: could not read file %s: %s\n", siteConfigFile, err)
			siteConfigs.UpdateSiteConfigError(errorConfigMap, errorNs, siteConfigFile, errorMsg)
			continue
		}

		siteConfig := siteConfigs.SiteConfig{}
		err = yaml.Unmarshal(fileData, &siteConfig)
		if err != nil {
			errorMsg := fmt.Sprintf("Error: could not parse %s as yaml: %s\n", siteConfigFile, err)
			siteConfigs.UpdateSiteConfigError(errorConfigMap, errorNs, siteConfigFile, errorMsg)
			continue
		}

		// overwrite the default extraManifestOnly with optional command line argument
		if *extraManifestOnly {
			for id := range siteConfig.Spec.Clusters {
				siteConfig.Spec.Clusters[id].ExtraManifestOnly = *extraManifestOnly
			}
		}

		clusters, err := scBuilder.Build(siteConfig)
		if err != nil {
			errorMsg := fmt.Sprintf("Error: could not build the entire SiteConfig defined by %s: %s", siteConfigFile, err)
			siteConfigs.UpdateSiteConfigError(errorConfigMap, errorNs, siteConfigFile, errorMsg)
			continue
		}

		for cluster, crs := range clusters {
			for _, crIntf := range crs {
				cr, err := yaml.Marshal(crIntf)
				if err != nil {
					errorMsg := fmt.Sprintf("Error: could not marshal generated cr by %s: %s %s", siteConfigFile, crIntf, err)
					siteConfigs.UpdateSiteConfigError(errorConfigMap, errorNs, siteConfigFile, errorMsg)
				} else {
					// write to file when out dir is provided, otherwise write to standard output
					if *outPath != siteConfigs.UnsetStringValue {
						crName := crIntf.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
						crKind := crIntf.(map[string]interface{})["kind"].(string)
						filePath := cluster + "_" + strings.ToLower(crKind) + "_" + crName + siteConfigs.FileExt
						err := siteConfigs.WriteFile(filePath, *outPath, cr)
						if err != nil {
							log.Printf("Error: could not write file %s: %s\n", *outPath+"/"+filePath, err)
						}
					} else {
						fmt.Print(string(siteConfigs.Separator))
						fmt.Println(string(cr))
					}
				}
			}
		}
	}

	if (errorConfigMap.(map[string]interface{})["data"] != nil) {
	    siteConfigs.PrintSiteConfigError(errorConfigMap, errorNs)
	}
}
