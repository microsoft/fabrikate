## Component Object Model

A deployment definition in Fabrikate is specified via one or more component.yaml/json files. Each of these file contains a
specification of a component with the following schema:

- `name`: A free form text name for this component. This name is used to refer to the component in [config specifications](./config.md).

- `generator`: Method used to generate the resource manifests for this particular component. Currently, `static` (file based), `helm` (helm based), and `component` are supported values.

- `source`: The source for this component. This can be a URL in the case of remote components or a local path to specify a local filesystem component.

- `method`: The method by which this component is sourced. Currently, only `git` and `local` are supported values.

- `path`: For some components, like ones generated with `helm`, the desired target of the component might not be located at the root of the repo. Path enables you to specify the relative `path` to this target from the root of the `source`.

- `version`: For git `method` components, this specifies a specific commit SHA hash that the component should be locked to, enabling you to lock the component to a consistent version.

- `branch`: For git `method` components, this specifies the branch that should be checked out after the git `source` is cloned.

- `hooks`: Hooks enable you to execute one or more shell commands before or after the following component lifecycle events: `before-install`, `before-generate`, `after-install`, `after-generate`.

- `repositories`: A field of key/value pairs consisting of a set of helm repositories that should be added.

- `subcomponents`: Zero or more subcomponents that define how to build the resource manifests that make up this component. These subcomponents are components themselves and have exactly the same schema as above.

## Examples

### Prometheus Grafana

This [component specification](https://github.com/timfpark/fabrikate-prometheus-grafana) generates static manifests for the `grafana` and `prometheus` namespaces and then remotely sources two helm charts for prometheus and grafana respectively.

```yaml
name: "prometheus-grafana"
generator: "static"
path: "./manifests"
subcomponents:
  - name: "grafana"
    generator: "helm"
    source: "https://github.com/helm/charts"
    method: "git"
    path: "stable/grafana"
  - name: "prometheus"
    generator: "helm"
    source: "https://github.com/helm/charts"
    method: "git"
    path: "stable/prometheus"
```

### Istio

This [component specification](https://github.com/evanlouie/fabrikate-istio) utilizes hooks to download and unpack an Istio release and then reference it with a local path.

```yaml
name: istio
generator: helm
path: "./tmp/istio-1.1.2/install/kubernetes/helm/istio"
hooks:
  before-install:
    - curl -Lv https://github.com/istio/istio/releases/download/1.1.2/istio-1.1.2-linux.tar.gz -o istio.tar.gz
    - mkdir -p tmp
    - tar xvf istio.tar.gz -C tmp
  after-install:
    - rm istio.tar.gz
subcomponents:
  - name: istio-namespace
    generator: static
    path: ./manifests
  - name: istio-crd # 1.1 split out CRDs to seperate chart
    generator: helm
    path: "./tmp/istio-1.1.2/install/kubernetes/helm/istio-init"
```

### Jaeger

This [component specification](https://github.com/bnookala/fabrikate-jaeger) is specified in JSON and also utilizes a `repositories` field to add the incubator repo to helm.

```json
{
  "name": "fabrikate-jaeger",
  "generator": "static",
  "path": "./manifests",
  "subcomponents": [
    {
      "name": "jaeger",
      "generator": "helm",
      "repositories": {
        "incubator": "https://kubernetes-charts-incubator.storage.googleapis.com/"
      },
      "source": "https://github.com/helm/charts",
      "method": "git",
      "path": "incubator/jaeger"
    }
  ]
}
```
