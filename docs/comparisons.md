# Fabrikate vs. Other Tools

This document serves to compare Fabrikate against available open source tools for managing deployments in a GitOps workflow.

These charts are in no way exhaustive or comprehensive, but serve as a starting point to help decide on a tool to use.

### Description
A short blurb on each tool to describe how each operates at a high level.

| [Fabrikate](https://github.com/microsoft/fabrikate) | [Helm Operator](https://github.com/fluxcd/helm-operator) | [Kustomize](https://github.com/kubernetes-sigs/kustomize) |
|:--|:--|:--|
| Fabrikate helps make operating Kubernetes clusters with a GitOps workflow more productive. It allows you to write DRY resource definitions and configuration for multiple environments while leveraging the broad Helm chart ecosystem, capture higher level definitions into abstracted and shareable components. <br><br> When used with a GitOps deployment tool like [Flux](https://github.com/fluxcd/flux), the cluster should reflect the state of the source repository. <br><br> Fabrikate can be run independently of a Kubernetes cluster. In an example workflow, fabrikate would be run as a part of a CI/CD pipeline and the resulting manifests would be synched to a cluster via a GitOps operator. | The Helm Operator is a Kubernetes operator, allowing one to declaratively manage Helm chart releases. <br><br> The desired state of a Helm release is described through a Kubernetes Custom Resource named HelmRelease. Based on the creation, mutation or removal of a HelmRelease resource in the cluster, Helm actions are performed by the operator. <br><br>  The Helm Operator runs on a kubernetes cluster in conjunction with [Flux](https://github.com/fluxcd/flux).| TODO |

### Abstraction
Descriptions on how different tools can be used to configure releases for multiple environments or clusters.
| [Fabrikate](https://github.com/microsoft/fabrikate) | [Helm Operator](https://github.com/fluxcd/helm-operator) | [Kustomize](https://github.com/kubernetes-sigs/kustomize) |
|:--|:--|:--|
| Fabrikate manages both static Kubernetes manifests or Helm charts by wrapping them into [Components](https://github.com/microsoft/fabrikate/blob/develop/docs/component.md). Helm charts in a component can have their values set via config .yaml or .json values that fabrikate will then generate static manifests. <br><br> A single fabrikate component can be mapped to multiple config files so that different sets of releases can be created and managed. <br><br> Multiple components can be combined into a [High Level Definitions](https://github.com/microsoft/bedrock/blob/master/docs/gitops-pipeline.md#deep-dive-high-level-definitions) for ease of management and repeatability. | The Helm Operator makes use of Helm Release CRDs to manage releases and overwrite default values of a helm chart. Multiple Helm Releases can point to the same Helm Chart and parameterize the values for each release. | TODO |

### Deployment Requirements and Components
Artifacts and dependencies to utilize each tool.
| [Fabrikate](https://github.com/microsoft/fabrikate) | [Helm Operator](https://github.com/fluxcd/helm-operator) | [Kustomize](https://github.com/kubernetes-sigs/kustomize) |
|:--|:--|:--|
| A repository for Helm Charts, a repository for Components/High Level Definitions, a repository for generated manifests. <br><br> Optionally these could all be the same repository, but this is not recommended. <br><br> Additionally a CI/CD pipeline should be used with fabrikate to generate the final manifests. | A repository for Helm Charts and another repository for Helm Releases. <br><br> Optionally, these could be the same repository. | TODO |

### Deployment Workflow
An example workflow for deploying the latest version application to a cluster.
| [Fabrikate](https://github.com/microsoft/fabrikate) | [Helm Operator](https://github.com/fluxcd/helm-operator) | [Kustomize](https://github.com/kubernetes-sigs/kustomize) |
|:--|:--|:--|
| Once an application image is built and published to a container repository, then the `image.tag` config value in the Component/HLD repository needs to be updated to the latest version, either manually or as an automated step after publishing the image in a CI/CD pipeline. Once the version is set, then the fabrikate generation pipeline should be triggered to generate the latest manifests with the updated image tag. Once the final manifests are generated, the GitOps operator should sync the changes onto the cluster. | Once an application image is built and published to a container repository, then the associated Helm Release needs to have the `image.tag` value updated to the latest version, either manually or as an automated step after publishing the image in a CI/CD pipeline. Once the Helm Release has the latest values, then the Helm Operator will deploy the latest image to the repository. | TODO |

### Rollback Workflow
Steps to rollback a deployment.
| [Fabrikate](https://github.com/microsoft/fabrikate) | [Helm Operator](https://github.com/fluxcd/helm-operator) | [Kustomize](https://github.com/kubernetes-sigs/kustomize) |
|:--|:--|:--|
| If using a GitOps operator like Flux, then the source manifest repository needs to be rolled back to the previous version. The best way to do so is to rollback the last configuration change commit in the Component/HLD repository. The resulting change should trigger the fabrikate pipeline to generate the last version of manifests again. | The latest changes in the Helm Release repository would need to be rolled back. The Helm Operator should then revert the changes on the next release. | TODO |

### Observability
| [Fabrikate](https://github.com/microsoft/fabrikate) | [Helm Operator](https://github.com/fluxcd/helm-operator) | [Kustomize](https://github.com/kubernetes-sigs/kustomize) |
|:--|:--|:--|
| The greatest asset of working with the fabrikate workflow is that the GitOps source repository will reflect the exact declarative state that the cluster is expected to be in. <br><br> The [fabrikate-definitions](https://github.com/microsoft/fabrikate-definitions) repository has sets of preconfigured observability HLD components that can easily be added to any deployment. | To see the expected state of a cluster, a user must connect to a cluster and utilize either the Helm or the Kubernetes CLIs. | TODO |

### Getting a Cluster Configured
| [Fabrikate](https://github.com/microsoft/fabrikate) | [Helm Operator](https://github.com/fluxcd/helm-operator) | [Kustomize](https://github.com/kubernetes-sigs/kustomize) |
|:--|:--|:--|
| The chosen GitOps operator needs to be added to the cluster and configured to the final generated manifest repository. <br><br> The [bedrock](https://github.com/microsoft/bedrock) and [bedrock-cli](https://github.com/microsoft/bedrock-cli) projects can be leveraged to configure clusters as well as supporting pipelines in an automated and repeatable pattern. | The Flux operator and Helm Operator must be installed onto a cluster then configured to the appointed Helm Release repository.| TODO |

### Helm version support
| [Fabrikate](https://github.com/microsoft/fabrikate) | [Helm Operator](https://github.com/fluxcd/helm-operator) | [Kustomize](https://github.com/kubernetes-sigs/kustomize) |
|:--|:--|:--|
| Both Helm 2 and 3 are supported. <br><br> **Note** The latest releases of fabrikate support Helm v3. To use fabrikate with Helm2, either use an [older release](https://github.com/microsoft/fabrikate/wiki/Roadmap-Updates#release-v1-and-v2) or build a binary off the [helm2](https://github.com/microsoft/fabrikate/tree/helm2) branch. | Both Helm 2 and 3 are supported | TODO |

### Support
| [Fabrikate](https://github.com/microsoft/fabrikate) | [Helm Operator](https://github.com/fluxcd/helm-operator) | [Kustomize](https://github.com/kubernetes-sigs/kustomize) |
|:--|:--|:--|
| Fabrikate is supported as a part of the wider Open Source [bedrock](https://github.com/microsoft/bedrock) project. [Please join us on Slack.](https://github.com/microsoft/bedrock#community) | The Helm Operator is part of Flux, a current sandbox project in CNCF. [More information can be found on the Helm Operator project page.](https://github.com/fluxcd/helm-operator#getting-help) | TODO |


