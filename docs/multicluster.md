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

## Add and configure `cloud-native` resource definitions
1. Add `cloud-native` component
    - `$ fab add cloud-native --source https://github.com/timfpark/fabrikate-cloud-native`
2. Set common storage class configuration for all environments in Azure:
    - ` $ fab set --subcomponent cloud-native.elasticsearch data.persistence.storageClass="managed-premium" master.persistence.storageClass="managed-premium"`
3. Set public IP address for west region 
    - ` $ fab set --environment west --subcomponent cloud-native.istio gateways.istio-ingressgateway.loadBalancerIP=40.65.120.17 gateways.istio-ingressgateway.serviceAnnotations."service.beta.kubernetes.io/azure-load-balancer-resource-group"="mycluster-west-rg"`
4. Set public IP address for central region 
    - ` $ fab set --environment west --subcomponent cloud-native.istio gateways.istio-ingressgateway.loadBalancerIP=45.63.429.29 gateways.istio-ingressgateway.serviceAnnotations."service.beta.kubernetes.io/azure-load-balancer-resource-group"="mycluster-central-rg"`
5. Set public IP address for east region 
    - ` $ fab set --environment west --subcomponent cloud-native.istio gateways.istio-ingressgateway.loadBalancerIP=59.32.793.49 gateways.istio-ingressgateway.serviceAnnotations."service.beta.kubernetes.io/azure-load-balancer-resource-group"="mycluster-east-rg"`

## Add `bookinfo` application resource definitions
1. Add Istio `bookinfo` example
    - `$ fab add bookinfo --source https://github.com/sarath-p/fabrikate-bookinfo`

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

## Validate `bookinfo` application
1. To test the `bookinfo` app, open a web browser to the public IP address configured for each cluster above.
2. To test the Azure Traffic Manager routing, open a browser to the Traffic Manager DNS name that is configured in [Bedrock](https://github.com/Microsoft/bedrock) [`azure-multi-cluster`](https://github.com/Microsoft/bedrock/tree/master/cluster/environments/azure-multiple-clusters) environment.

## Validate `cloud-native` stack
The user interfaces for the `cloud-native` services are not exposed publicly via an external ip address. To access the `cloud-native`` user interfaces, use the kubectl [port-forward](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#port-forward) command. This command creates a secure connection between a local port on your client machine and the relevant pod in your AKS cluster. Here are some examples:

1. Access `Grafana` analytics and monitoring dashboards 
    - `kubectl -n grafana port-forward $(kubectl -n grafana  get pod -l app=grafana -o jsonpath='{.items[0].metadata.name}') 3000:3000`
2. Access `Prometheus` metrics
    - `kubectl -n prometheus port-forward $(kubectl -n prometheus get pod -l app=prometheus -o jsonpath='{.items[0].metadata.name}') 9090:9090`
3. Access `Jaeger` tracing
    - `kubectl port-forward -n jaeger $(kubectl get pod -n jaeger -l app=jaeger -o jsonpath='{.items[0].metadata.name}') 16686:16686`

## References

1. The output of this example using the above steps is available in git repo [`fabrikate-cloud-native-bookinfo`](https://github.com/sarath-p/fabrikate-cloud-native-bookinfo).
2. [Istio Bookinfo Application](https://istio.io/docs/examples/bookinfo/)