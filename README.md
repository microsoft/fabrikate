# Fabrikate

[![Build Status][azure-devops-build-status]][azure-devops-build-link]
[![Go Report Card][go-report-card-badge]][go-report-card]

Fabrikate helps make operating Kubernetes clusters with a
[GitOps](https://www.weave.works/blog/gitops-operations-by-pull-request)
workflow more productive. It allows you to write
[DRY](https://en.wikipedia.org/wiki/Don%27t_repeat_yourself) resource
definitions and configuration for multiple environments while leveraging the
broad [Helm chart ecosystem](https://github.com/helm/charts), capture higher
level definitions into abstracted and shareable components, and enable a
[GitOps](https://www.weave.works/blog/gitops-operations-by-pull-request)
deployment workflow that both simplifies and makes deployments more auditable.

In particular, Fabrikate simplifies the frontend of the GitOps workflow: it
takes a high level description of your deployment, a target environment
configuration (eg. `qa` or `prod`), and renders the Kubernetes resource
manifests for that deployment utilizing templating tools like
[Helm](https://helm.sh). It is intended to run as part of a CI/CD pipeline such
that with every commit to your Fabrikate deployment definition triggers the
generation of Kubernetes resource manifests that an in-cluster GitOps pod like
[Weaveworks' Flux](https://github.com/weaveworks/flux) watches and reconciles
with the current set of applied resource manifests in your Kubernetes cluster.

## Getting Started

First, install the latest `fab` cli on your local machine from
[our releases](https://github.com/microsoft/fabrikate/releases), unzipping the
appropriate binary and placing `fab` in your path. The `fab` cli tool, `helm`,
and `git` are the only tools you need to have installed.

Let's walk through building an example Fabrikate definition to see how it works
in practice. First off, let's create a directory for our cluster definition:

```sh
$ mkdir mycluster
$ cd mycluster
```

The first thing I want to do is pull in a common set of observability and
service mesh platforms so I can operate this cluster. My organization has
settled on a
[cloud-native](https://github.com/microsoft/fabrikate-definitions/tree/master/definitions/fabrikate-cloud-native)
stack, and Fabrikate makes it easy to leverage reusable stacks of infrastructure
like this:

```sh
$ fab add cloud-native --source https://github.com/microsoft/fabrikate-definitions --path definitions/fabrikate-cloud-native
```

Since our directory was empty, this creates a component.yaml file in this
directory:

```yaml
name: mycluster
subcomponents:
  - name: cloud-native
    type: component
    source: https://github.com/microsoft/fabrikate-definitions
    method: git
    path: definitions/fabrikate-cloud-native
    branch: master
```

A Fabrikate definition, like this one, always contains a `component.yaml` file
in its root that defines how to generate the Kubernetes resource manifests for
its directory tree scope.

The `cloud-native` component we added is a remote component backed by a git repo
[fabrikate-cloud-native](https://github.com/microsoft/fabrikate-definitions/tree/master/definitions/fabrikate-cloud-native).
Fabrikate definitions use remote definitions like this one to enable multiple
deployments to reuse common components (like this cloud-native infrastructure
stack) from a centrally updated location.

Looking inside this component at its own root `component.yaml` definition, you
can see that it itself uses a set of remote components:

```yaml
name: "cloud-native"
generator: "static"
path: "./manifests"
subcomponents:
  - name: "elasticsearch-fluentd-kibana"
    source: "../fabrikate-elasticsearch-fluentd-kibana"
  - name: "prometheus-grafana"
    source: "../fabrikate-prometheus-grafana"
  - name: "istio"
    source: "../fabrikate-istio"
  - name: "kured"
    source: "../fabrikate-kured"
```

Fabrikate recursively iterates component definitions, so as it processes this
lower level component definition, it will in turn iterate the remote component
definitions used in its implementation. Being able to mix in remote components
like this makes Fabrikate deployments composable and reusable across
deployments.

Let's look at the component definition for the
[elasticsearch-fluentd-kibana component](https://github.com/microsoft/fabrikate-definitions/tree/master/definitions/fabrikate-elasticsearch-fluentd-kibana):

```json
{
  "name": "elasticsearch-fluentd-kibana",
  "generator": "static",
  "path": "./manifests",
  "subcomponents": [
    {
      "name": "elasticsearch",
      "generator": "helm",
      "source": "https://github.com/helm/charts",
      "method": "git",
      "path": "stable/elasticsearch"
    },
    {
      "name": "elasticsearch-curator",
      "generator": "helm",
      "source": "https://github.com/helm/charts",
      "method": "git",
      "path": "stable/elasticsearch-curator"
    },
    {
      "name": "fluentd-elasticsearch",
      "generator": "helm",
      "source": "https://github.com/helm/charts",
      "method": "git",
      "path": "stable/fluentd-elasticsearch"
    },
    {
      "name": "kibana",
      "generator": "helm",
      "source": "https://github.com/helm/charts",
      "method": "git",
      "path": "stable/kibana"
    }
  ]
}
```

First, we see that components can be defined in JSON as well as YAML (as you
prefer).

Secondly, we see that that this component generates resource definitions. In
particular, it will emit a set of static manifests from the path `./manifests`,
and generate the set of resource manifests specified by the inlined
[Helm templates](https://helm.sh/) definitions as it it iterates your deployment
definitions.

With generalized helm charts like the ones used here, its often necessary to
provide them with configuration values that vary by environment. This component
provides a reasonable set of defaults for its subcomponents in
`config/common.yaml`. Since this component is providing these four logging
subsystems together as a "stack", or preconfigured whole, we can provide
configuration to higher level parts based on this knowledge:

```yaml
config:
subcomponents:
  elasticsearch:
    namespace: elasticsearch
    injectNamespace: true
    config:
      client:
        resources:
          limits:
            memory: "2048Mi"
  elasticsearch-curator:
    namespace: elasticsearch
    injectNamespace: true
    config:
      cronjob:
        successfulJobsHistoryLimit: 0
      configMaps:
        config_yml: |-
          ---
          client:
            hosts:
              - elasticsearch-client.elasticsearch.svc.cluster.local
            port: 9200
            use_ssl: False
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

This `common` configuration, which applies to all environments, can be mixed
with more specific configuration. For example, let's say that we were deploying
this in Azure and wanted to utilize its `managed-premium` SSD storage class for
Elasticsearch, but only in `azure` deployments. We can build an `azure`
configuration that allows us to do exactly that, and Fabrikate has a convenience
function called `set` that enables to do exactly that:

```
$ fab set --environment azure --subcomponent cloud-native.elasticsearch data.persistence.storageClass="managed-premium" master.persistence.storageClass="managed-premium"
```

This creates a file called `config/azure.yaml` that looks like this:

```yaml
subcomponents:
  cloud-native:
    subcomponents:
      elasticsearch:
        config:
          data:
            persistence:
              storageClass: managed-premium
          master:
            persistence:
              storageClass: managed-premium
```

Naturally, an observability stack is just the base infrastructure we need, and
our real goal is to deploy a set of microservices. Furthermore, let's assume
that we want to be able to split the incoming traffic for these services between
`canary` and `stable` tiers with [Istio](https://istio.io) so that we can more
safely launch new versions of the service.

There is a Fabrikate component for that as well called
[fabrikate-istio-service](https://github.com/microsoft/fabrikate-definitions/tree/master/definitions/fabrikate-istio)
that we'll leverage to add this service, so let's do just that:

```
$ fab add simple-service --source https://github.com/microsoft/fabrikate-definitions --path definitions/fabrikate-istio
```

This component creates these traffic split services using the config applied to
it. Let's create a `prod` config that does this for a `prod` cluster by creating
`config/prod.yaml` and placing the following in it:

```yaml
subcomponents:
  simple-service:
    namespace: services
    config:
      gateway: my-ingress.istio-system.svc.cluster.local
      service:
        dns: simple.mycompany.io
        name: simple-service
        port: 80
      configMap:
        PORT: 80
      tiers:
        canary:
          image: "timfpark/simple-service:441"
          replicas: 1
          weight: 10
          port: 80
          resources:
            requests:
              cpu: "250m"
              memory: "256Mi"
            limits:
              cpu: "1000m"
              memory: "512Mi"

        stable:
          image: "timfpark/simple-service:440"
          replicas: 3
          weight: 90
          port: 80
          resources:
            requests:
              cpu: "250m"
              memory: "256Mi"
            limits:
              cpu: "1000m"
              memory: "512Mi"
```

This defines a service that is exposed on the cluster via a particular gateway
and dns name and port. It also defines a traffic split between two backend
tiers: `canary` (10%) and `stable` (90%). Within these tiers, we also define the
number of replicas and the resources they are allowed to use, along with the
container that is deployed in them. Finally, it also defines a ConfigMap for the
service, which passes along an environmental variable to our app called `PORT`.

From here we could add definitions for all of our microservices in a similar
manner, but in the interest of keeping this short, we'll just do one of the
services here.

With this, we have a functionally complete Fabrikate definition for our
deployment. Let's now see how we can use Fabrikate to generate resource
manifests for it.

First, let's install the remote components and helm charts:

```sh
$ fab install
```

This installs all of the required components and charts locally and we can now
generate the manifests for our deployment with:

```sh
$ fab generate prod azure
```

This will iterate through our deployment definition, collect configuration
values from `azure`, `prod`, and `common` (in that priority order) and generate
manifests as it descends breadth first. You can see the generated manifests in
`./generated/prod-azure`, which has the same logical directory structure as your
deployment definition.

Fabrikate is meant to used as part of a CI / CD pipeline that commits the
generated manifests checked into a repo so that they can be applied from a pod
within the cluster like [Flux](https://github.com/weaveworks/flux), but if you
have a Kubernetes cluster up and running you can also apply them directly with:

```sh
$ cd generated/prod-azure
$ kubectl apply --recursive -f .
```

This will cause a very large number of containers to spin up (which will take
time to start completely as Kubernetes provisions persistent storage and
downloads the containers themselves), but after three or four minutes, you
should see the full observability stack and Microservices running in your
cluster.

## Documentation

We have complete details about how to use and contribute to Fabrikate in these
documentation items:

- [Component Definitions](./docs/component.md)
- [Config Definitions](./docs/config.md)
- [Command Reference](./docs/commands.md)
- [Authentication / Personal Access Tokens (PAT) / `access.yaml`](./docs/auth.md)
- [Contributing](./docs/contributing.md)

## Community

[Please join us on Slack](https://join.slack.com/t/bedrockco/shared_invite/enQtNjIwNzg3NTU0MDgzLTdiZGY4ZTM5OTM4MWEyM2FlZDA5MmE0MmNhNTQ2MGMxYTY2NGYxMTVlZWFmODVmODJlOWU0Y2U2YmM1YTE0NGI)
for discussion and/or questions.

## Bedrock

We maintain a sister project called
[Bedrock](https://github.com/microsoft/bedrock). Bedrock provides automata that
make operationalizing Kubernetes clusters with a GitOps deployment workflow
easier, automating a
[GitOps](https://www.weave.works/blog/gitops-operations-by-pull-request)
deployment model leveraging [Flux](https://github.com/weaveworks/flux), and
provides automation for building a CI/CD pipeline that automatically builds
resource manifests from Fabrikate defintions.

<!-- refs -->

[azure-devops-build-status]:
  https://tpark.visualstudio.com/fabrikate/_apis/build/status/microsoft.fabrikate?branchName=master
[azure-devops-build-link]:
  https://tpark.visualstudio.com/fabrikate/_build/latest?definitionId=35&branchName=master
[go-report-card]: https://goreportcard.com/report/github.com/microsoft/fabrikate
[go-report-card-badge]:
  https://goreportcard.com/badge/github.com/microsoft/fabrikate
