# Generate and Deploy Kubenetes Resource Manifests using Fabrikate and Bedrock

This short guide shows a self-contained example of generating resource manifests for multicluster [`cloud-native`](https://github.com/timfpark/fabrikate-cloud-native/) stack using [Frabrikate](https://github.com/Microsoft/fabrikate) and deploying with [BedRock](https://github.com/Microsoft/bedrock) GitOps. It just takes few of minutes to get set up. By the end you will have [`cloud-native`](https://github.com/timfpark/fabrikate-cloud-native/) running in your cluster in each region.


## Prerequisites

1. Deploy three AKS clusters with Traffic Manager routing in Azure using [Bedrock](https://github.com/Microsoft/bedrock) [`azure-multi-cluster`](https://github.com/Microsoft/bedrock/tree/master/cluster/environments/azure-multiple-clusters) environment.
2. Get public ip address from each region/cluster from the above environment
3. Get Azure resoirce group name of each region/clluster from the above environment
4. Configured following cluster variables as specified:
    - `gitops_west_path` = "generated/prod-west"
    - `gitops_central_path` = "generated/prod-central"
    - `gitops_east_path`= "generated/prod-east"
4. Open terminal/WSL shell in cloned directory root of the git repo that was configured for `gitops_ssh_url` in cluster deployment

## Create Resource Definitions
1. Add `cloud-native` component
    - `$ fab add cloud-native --source https://github.com/timfpark/fabrikate-cloud-native`
2. Add common configuration for all environments in Azure:
    - ` $ fab set --subcomponent cloud-native.elasticsearch data.persistence.storageClass="managed-premium" master.persistence.storageClass="managed-premium"`
3. Set public IP address for west region 
    - ` $ fab set --environment west --subcomponent cloud-native.istio gateways.istio-ingressgateway.loadBalancerIP=40.65.120.17 gateways.istio-ingressgateway.serviceAnnotations."service.beta.kubernetes.io/azure-load-balancer-resource-group"="mycluster-west-rg"`
4. Set public IP address for central region 
    - ` $ fab set --environment west --subcomponent cloud-native.istio gateways.istio-ingressgateway.loadBalancerIP=45.63.429.29 gateways.istio-ingressgateway.serviceAnnotations."service.beta.kubernetes.io/azure-load-balancer-resource-group"="mycluster-central-rg"`
5. Set public IP address for east region 
    - ` $ fab set --environment west --subcomponent cloud-native.istio gateways.istio-ingressgateway.loadBalancerIP=59.32.793.49 gateways.istio-ingressgateway.serviceAnnotations."service.beta.kubernetes.io/azure-load-balancer-resource-group"="mycluster-east-rg"`

## Generate Resource Manifests
* Install remote components
    - `$ fab install`
* Generate resource manafetsts for `prod` deployment with `west` region specific configuration
    - `$ fab generate prod west`
* Generate resource manafetsts for `prod` deployment with `central` region specific configuration
    - `$ fab generate prod central`
* Generate resource manafetsts for `prod` deployment with `east` region specific configuration
    - `$ fab generate prod east`

## Push Manifests to Git repo
* `$ git add generated/`
* `$ git commit -m "adding fab generated resource manifests."`
* `$ git push`

## Test
1. Deploy [Bookinfo application from Istio exampes](
https://istio.io/docs/examples/bookinfo/) to test the deployment thorugh a working application.

