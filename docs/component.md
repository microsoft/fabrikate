# Component Object Model

A deployment definition in Fabrikate is specified via one or more
component.yaml/json files. Each of these file contains a specification of a
component with the following schema:

- `name`: A free form text name for this component. This name is used to refer
  to the component in [config specifications](./config.md).

- `type`: Method used generate the manifests for this particular component.
  Currently, `static` (static manifest based), `helm` (helm based), and
  `component` (default) are supported values.

  - if `type: component`: the component itself does not contain any manifests to
    generate, but is a container for other components.
  - if `type: helm`: the component will use `helm template` to materialize the
    component using the specified config under `config` as the `values.yaml`
    file.
  - if `type: static`: the component holds raw kubernetes manifest files in
    `path`, these manifests will be copied to the generated output.

- `method`: The method by which this component is sourced. Currently, only
  `git`, `helm`, and `local` are supported values.

  - if `method: git`: Tells fabrikate to `git clone <source>`.
  - if `method: helm`: Tells fabrikate to `helm fetch <source>/<path>` from the
    `source` helm repo. Essentially:
    `helm repo add foo <my_helm_repository> && helm fetch foo/<path>`
  - if `method: local`: Tells fabrikate to use the host filesystem as a means to
    find the component.

- `source`: The source for this component.

  - if `method: git`: A URL for a Git repository (the url you would call
    `git clone` on).
  - if `method: helm`: A URL to a helm repository (the url you would call
    `helm repo add` on).
  - if `method: local`: A local path to specify a local filesystem component.

- `path`: For some components, like ones generated with `helm`, the desired
  target of the component might not be located at the root of the repo. Path
  enables you to specify the relative `path` to this target from the root of the
  `source`.

  - if `method: git`: the subdirectory of the component in the git repo
    specified in `source`.
  - if `method: helm`: the name of the chart to install the repo specified in
    `source`.
  - if `method: local`: the subdirectory on host filesystem where the component
    is located.

- `version`: For git `method` components, this specifies a specific commit SHA
  hash that the component should be locked to, enabling you to lock the
  component to a consistent version.

  - if `method: git`: a specific commit to checkout from the repository.
  - if `method: helm`: noop
  - if `method: local`: noop

- `branch`: For git `method` components, this specifies the branch that should
  be checked out after the git `source` is cloned.

  - if `method: git`: a specific branch to checkout from the repository.
  - if `method: helm`: noop
  - if `method: local`: noop

- `hooks`: Hooks enable you to execute one or more shell commands before or
  after the following component lifecycle events: `before-install`,
  `before-generate`, `after-install`, `after-generate`.

- `subcomponents`: Zero or more subcomponents that define how to build the
  resource manifests that make up this component. These subcomponents are
  components themselves and have exactly the same schema as above.

## Examples

### Prometheus Grafana

This
[component specification](https://github.com/timfpark/fabrikate-prometheus-grafana)
generates static manifests for the `grafana` and `prometheus` namespaces and
then remotely sources two helm charts for prometheus and grafana respectively.

Grafana is sourced via `method: git`, so the entire repository is cloned and the
`path` is utilized to point to the location of the helm chart.

Prometheus is sourced via `method: helm`, so the incubator repo noted in
`source` is temporarily added to the hosts helm client and the chart noted in
`path` is `helm fetch`d from the `source` repo.

```yaml
name: "prometheus-grafana"
generator: "static"
path: "./manifests"
subcomponents:
  - name: "grafana"
    type: "helm"
    method: "git" # fetch the `source` via `git clone`
    source: "https://github.com/helm/charts" # source of stable helm repo
    path: "stable/grafana" # path in stable helm repo to chart
  - name: "prometheus"
    type: "helm"
    method: "helm" # fetch the `source` via `helm fetch`
    source: "https://kubernetes-charts.storage.googleapis.com" # url of helm repo the chart resides
    path: "prometheus" # name of chart in helm repository
```

### Istio

This [component specification](https://github.com/evanlouie/fabrikate-istio)
utilizes hooks to download and unpack an Istio release and then reference it
with a local path.

```yaml
name: istio
generator: helm
path: "./tmp/istio-1.1.2/install/kubernetes/helm/istio"
hooks:
  before-install:
    - |
      curl -Lv https://github.com/istio/istio/releases/download/1.1.2/istio-1.1.2-linux.tar.gz -o istio.tar.gz
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

This [component specification](https://github.com/bnookala/fabrikate-jaeger) is
specified in JSON and utilizes the `method: helm` feature to `helm fetch` the
jaeger chart from the incubator helm repo.

```json
{
  "name": "fabrikate-jaeger",
  "generator": "static",
  "path": "./manifests",
  "subcomponents": [
    {
      "name": "jaeger",
      "type": "helm",
      "method": "helm",
      "source": "https://kubernetes-charts-incubator.storage.googleapis.com/",
      "path": "jaeger"
    }
  ]
}
```
