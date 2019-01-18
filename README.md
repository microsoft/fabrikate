# fabrikate

Fabrikate makes GitOps devops for Kubernetes clusters easier. It allows you to write [DRY](https://en.wikipedia.org/wiki/Don%27t_repeat_yourself) resource definitions and configuration for multiple environments, capturing common resource definitions into abstracted and shareable components, and enabling a [GitOps](https://www.weave.works/blog/gitops-operations-by-pull-request) deployment workflow that both simplifies and makes deployments more auditable.

In particular, Fabrikate simplifies the frontend of the GitOps workflow: it takes a high level description of your deployment, a target environment (eg. `dev` or `prod`), and renders the Kubernetes resource manifests for that deployment. It is intended to run as part of your CI/CD pipeline such that with every commit to your deployment project triggers the generation of Kubernetes resource manifests that a tool like [Flux](https://github.com/weaveworks/flux) then automatically reconciles with the current state of your Kubernetes cluster.

## Getting Started

First, install the latest `fab` cli on your local machine from [our releases](https://github.com/Microsoft/fabrikate/releases), unzipping the appropriate binary for your operating system and placing `fab` in your path.  The `fab` cli tool, `docker`, and `git` are the only tools you need to have installed.  Any other tooling dependencies will be fetched via `docker` images and/or `git`.

Let's walk through an example project to see how Fabrikate works in practice.

```
$ git clone https://github.com/Microsoft/fabrikate
$ cd fabrikate/examples/getting-started
```

This directory is the root of a Fabrikate deployment project and contains a `component.json` file for the current component. A component in Fabrikate is the definition for building the Kubernetes resource definitions for its directory tree scope.

```
{
    "name": "microservices",
    "subcomponents": [
        {
            "name": "infra",
            "source": "./infra"
        },
        {
            "name": "services",
            "source": "./services"
        }
    ]
}
```

In this case, it defines a component called "microservices" that consists of two subcomponents, "infra" and "services".

Let's look at the `infra` directory.  This directory defines the common application infrastructure that all of the microservices in our deployment will use. 

```
{
    "name": "cloud-native-infra",
    "subcomponents": [
        {
            "name": "elasticsearch-fluentd-kibana",
            "source": "https://github.com/timfpark/fabrikate-elasticsearch-fluentd-kibana",
            "method": "git"
        },
        {
            "name": "prometheus-grafana",
            "source": "https://github.com/timfpark/fabrikate-prometheus-grafana",
            "method": "git"
        }
    ]
}
```

In this case, it will build resource manifests for both Elasticsearch / FluentD / Kibana (EFK) log management and Prometheus / Grafana metrics monitoring stacks using components in an external git repo.  Fabrikate enables linking out to components in an external repo like this as a way of sharing common subcomponents between deployment projects.

Looking at the [backing repo](https://github.com/timfpark/fabrikate-elasticsearch-fluentd-kibana) for the EFK component, we can see that it also defines a `component.json`:

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

But in this case, it defines subcomponents that are backed by a helm chart on the project repo. Fabrikate will clone the helm repo and generate the resource manifests from them.  The `elasticsearch-fluentd-kibana` component itself is a static component and has a number of manifests in the `./manifest` that will directly be included in the generated overall manifests.

With generalized helm charts like the ones used here, its often necessary to provide them with configuration values that vary by environment. This component provides a reasonable set of defaults for its subcomponents in `config/common.json`:

```
{
    "config": {},
    "subcomponents": {
        "elasticsearch": {
            "config": {
                "namespace": "elasticsearch",
                "data": {
                    "persistence": {
                        "storageClass": "default"
                    }
                },
                "master": {
                    "persistence": {
                        "storageClass": "default"
                    }
                }
            }
        },
        "fluentd-elasticsearch": {
            "config": {
                "elasticsearch": {
                    "host": "elasticsearch-client.elasticsearch.svc.cluster.local"
                },
                "namespace": "fluentd"
            }
        },
        "kibana": {
            "config": {
                "files": {
                    "kibana.yml": {
                        "elasticsearch.url": "http://elasticsearch-client.elasticsearch.svc.cluster.local:9200"
                    }
                },
                "namespace": "kibana"
            }
        }
    }
}
```

These values can be overridden by more specific environments (eg. a `prod.json` in this directory) or configuration higher in the directory tree for the deployment definition. Our example uses this to override the Elasticsearch `storageClass` to use `managed-premium` (Azure's SSD storage class) for the `prod` deployment in `infra/config/prod.json`: 

```
{
    "config": {},
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

```

With that explanation done, let's install the remote components by going to the `~/examples/getting-started` root of our definition:

```
$ fab install
```

With those installed, we can now generate the manifests for our deployment with:

```
$ fab generate prod
```

This will iterate through our deployment definition, collecting configuration values and generating manifests as it descends breadth first.  You can see the generated manifests in `./generated`, which has the same logical directory structure as your deployment definition.

These manifests are meant to be generated as part of a CI / CD pipeline and then applied from a daemon within the cluster like [Flux](https://github.com/weaveworks/flux), but if you have a Kubernetes cluster up and running you can also apply them directly with:

```
$ cd generated/prod
$ kubectl apply --recursive -f .
```

##  Contributing

This project welcomes contributions and suggestions.  Most contributions require you to agree to a Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us the rights to use your contribution. For details, visit https://cla.microsoft.com.

When you submit a pull request, a CLA-bot will automatically determine whether you need to provide a CLA and decorate the PR appropriately (e.g., label, comment). Simply follow the instructions provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.


