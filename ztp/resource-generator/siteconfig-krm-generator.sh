#!/usr/bin/env bash

siteConfigPath=/kustomize/plugin/ran.openshift.io/v1/siteconfig

EMS_ONLY=${extraManifestOnly:-"false"}
WRAP_IN_POLICY=${wrapInPolicy:-"true"}

resourceList=$(cat) # read the 'kind: ResourceList' from stdin
num=$(echo "$resourceList" | yq eval '.items | length')

echo $num >> log_file_site
echo $resourceList >> log_file_site
echo $(pwd) >> log_file_site
echo $(ls -al) >> log_file_site


for (( i=0; i<$num; i++ )); do
   # the container is executed with user nobody by default, write the resource to /tmp which is accessible by the user
   echo "$resourceList" | yq eval -o=y ".items[\"$i\"]" > resource_$i.yaml

   echo "Check if resource_$i.yaml exists" >> log_file_site
   echo $(ls) >> log_file_site

   resourceKind=$(yq eval '.kind' < resource_$i.yaml)
   echo "resourceKind: $resourceKind" >> log_file_site
   if [[ "$resourceKind" == "SiteConfig" ]]; then
       $siteConfigPath/SiteConfig \
         -manifestPath $siteConfigPath/extra-manifest \
         -extraManifestOnly=${EMS_ONLY} \
         resource_$i.yaml
       if [[ $? -ne 0 ]]; then
           echo "Error: failed to generate installation CRs $(pwd)"  >> /dev/stderr
           exit 1
       fi
   else
       echo "Log: resource of $resourceKind is not SiteConfig" >> log_file_site
   fi
done


