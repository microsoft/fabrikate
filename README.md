# fabrikate

Fabrikate is a tool to make operating Kubernetes clusters with a [GitOps](https://www.weave.works/blog/gitops-operations-by-pull-request) workflow more productive. It allows you to write [DRY](https://en.wikipedia.org/wiki/Don%27t_repeat_yourself) resource definitions and configuration for multiple environments, capture common resource definitions into abstracted and shareable components, and enable a [GitOps](https://www.weave.works/blog/gitops-operations-by-pull-request) deployment workflow that both simplifies and makes deployments more auditable.

In particular, Fabrikate simplifies the frontend of the GitOps workflow: it takes a high level description of your deployment, a target environment configuration (eg. `qa` or `prod`), and renders the Kubernetes resource manifests for that deployment. It is intended to run as part of a CI/CD pipeline such that with every commit to your high level deployment definition triggers the generation of Kubernetes resource manifests that an in-cluster GitOps pod like [Weaveworks' Flux](https://github.com/weaveworks/flux) watches and reconciles with the current set of applied resource manifests in your Kubernetes cluster.

## Getting Started

First, install the latest `fab` cli on your local machine from [our releases](https://github.com/Microsoft/fabrikate/releases), unzipping the appropriate binary and placing `fab` in your path.  The `fab` cli tool, `helm`, and `git` are the only tools you need to have installed.

Let's walk through an example high level definition to see how Fabrikate works in practice.

```
$ git clone https://github.com/timfpark/fabrikate-example
$ cd fabrikate-example
```

This example definition, like all Fabrikate definitions, contains a `component.yaml` file in its root that defines how to generate the Kubernetes resource manifests for its directory tree scope:

```
name: "my-microservices"
subcomponents:
  - name: "cloud-native"
    source: "https://github.com/timfpark/fabrikate-cloud-native"
    method: "git"
  - name: "services"
    source: "./services"
```

In this case, it defines a "my-microservices" component that defines the complete deployment of two subcomponents, `cloud-native` and `services`. `cloud-native` is a remote component backed by a git repo [fabrikate-cloud-native](https://github.com/timfpark/fabrikate-cloud-native). Fabrikate definitions use remote definitions like this one to enable multiple deployments to reuse common components like this cloud-native infrastructure stack from a centrally updated location.

Looking inside [fabrikate-cloud-native](https://github.com/timfpark/fabrikate-cloud-native/blob/master/component.yaml) for its own root `component.yaml` definition, you can see that it itself uses a set of remote components:

```
name: "cloud-native"
subcomponents:
  - name: "elasticsearch-fluentd-kibana"
    source: "https://github.com/timfpark/fabrikate-elasticsearch-fluentd-kibana"
    method: "git"
  - name: "prometheus-grafana"
    source: "https://github.com/timfpark/fabrikate-prometheus-grafana"
    method: "git"
  - name: "istio"
    source: "https://github.com/evanlouie/fabrikate-istio"
    method: "git"
  - name: "kured"
    source: "https://github.com/timfpark/fabrikate-kured"
    method: "git"
  - name: "jaeger"
    source: "https://github.com/bnookala/fabrikate-jaeger"
    method: "git"
```

Fabrikate recursively iterates component definitions, so as it processes this lower level component definition, it will in term iterate the remote component definitions used in its implementation.  Being able to mix in remote components like this makes Fabrikate deployments composable and reusable across deployments.

Looking at the component definition for the [elasticsearch-fluentd-kibana component](https://github.com/timfpark/fabrikate-elasticsearch-fluentd-kibana/blob/master/component.json):

```
{
    "name": "elasticsearch-fluentd-kibana",
    "generator": "static",
    "path": "./manifests",
    "subcomponents": [
        {
            "name": "elasticsearch",
            "generator": "helm",
            "repo": "https://github.com/helm/charts",
            "path": "stable/elasticsearch"
        },
        {
            "name": "elasticsearch-curator",
            "generator": "helm",
            "repo": "https://github.com/helm/charts",
            "path": "stable/elasticsearch-curator"
        },
        {
            "name": "fluentd-elasticsearch",
            "generator": "helm",
            "repo": "https://github.com/helm/charts",
            "path": "stable/fluentd-elasticsearch"
        },
        {
            "name": "kibana",
            "generator": "helm",
            "repo": "https://github.com/helm/charts",
            "path": "stable/kibana"
        }
    ]
}
```

First, we see that components can be defined in JSON as well as YAML (as you prefer).  

Secondly, we see that that this component generates resource definitions. In particular, it will emit a set of static manifests from the path `./manifests`, and generate the set of resource manifests specified by the inlined [Helm templates](https://helm.sh/) definitions as it it iterates your deployment definitions.

With generalized helm charts like the ones used here, its often necessary to provide them with configuration values that vary by environment. This component provides a reasonable set of defaults for its subcomponents in `config/common.yaml`.  Since this component is providing these four logging subsystems together as a "stack", or preconfigured whole, we can provide configuration to higher level parts based on this knowledge:

```
config:
subcomponents:
  elasticsearch:
    namespace: elasticsearch
    injectNamespace: true
    config:
  elasticsearch-curator:
    config:
      namespace: elasticsearch
      configMaps:
        config_yml: |-
          ---
          client:
            hosts:
              - elasticsearch-client.elasticsearch.svc.cluster.local
            port: 9200
            use_ssl: True
  fluentd-elasticsearch:
    namespace: fluentd
    injectNamespace: true
    config:
      elasticsearch:
        host: "elasticsearch-client.elasticsearch.svc.cluster.local"
  kibana:
    namespace: kibana
    injectNamespace: true
    config:
      files:
        kibana.yml:
          elasticsearch.url: "http://elasticsearch-client.elasticsearch.svc.cluster.local:9200"
```

These values can be overridden by more specific environments (eg. by adding a `azure.json` in this directory) or configuration higher in the directory tree for the deployment definition. Our example uses this to override the Elasticsearch `storageClass` to use the `managed-premium` (Azure's SSD storage class) for the `azure` configuration in `config/azure.json`: 

```
{
    "subcomponents": {
        "cloud-native": {
            "subcomponents": {
                "elasticsearch-fluentd-kibana": {
                    "config": {},
                    "subcomponents": {
                        "elasticsearch": {
                            "config": {
                                "data": {
                                    "persistence": {
                                        "storageClass": "managed-premium"
                                    }
                                },
                                "master": {
                                    "persistence": {
                                        "storageClass": "managed-premium"
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}
```

Given this overview of how the parts fit together in this example, let's see how we can use Fabrikate to generate resource manifests for this deployment.

First, let's install the remote components and helm charts:

```
$ fab install
```

With those installed, we can now generate the manifests for our deployment with:

```
$ fab generate azure
```

This will iterate through our deployment definition, collect configuration values and generate manifests as it descends breadth first.  You can see the generated manifests in `./generated`, which has the same logical directory structure as your deployment definition.

These manifests are meant to be generated as part of a CI / CD pipeline and then applied from a daemon within the cluster like [Flux](https://github.com/weaveworks/flux), but if you have a Kubernetes cluster up and running you can also apply them directly with:

```
$ cd generated/azure
$ kubectl apply --recursive -f .
```

## Bedrock

As one final note, we have also open sourced a project that helps you operationalize Kubernetes clusters with a GitOps deployment pipeline called [Bedrock](https://github.com/Microsoft/bedrock).  Bedrock provides automation for creating Kubernetes clusters, installs a [GitOps](https://www.weave.works/blog/gitops-operations-by-pull-request) deployment model leveraging Flux, and provides automation for building a CI/CD pipeline that automatically builds resource manifests from high level definitions like the example one we have been considering here.

##  Contributing

This project welcomes contributions and suggestions.  Most contributions require you to agree to a Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us the rights to use your contribution. For details, visit https://cla.microsoft.com.

When you submit a pull request, a CLA-bot will automatically determine whether you need to provide a CLA and decorate the PR appropriately (e.g., label, comment). Simply follow the instructions provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.
