# Fabrikate-Jaeger

This [fabrikate](http://github.com/microsoft/fabrikate) stack installs Jaeger on your cluster, with a provided "production" configuration.

### Requirements

- The [fabrikate 0.2.3](http://github.com/microsoft/fabrikate/releases) cli tool installed locally
- The [helm](https://github.com/helm/helm/releases) cli tool installed locally
- The kubectl cli tool installed locally

### Setup

Make sure your helm incubator repository is pointed at https://kubernetes-charts-incubator.storage.googleapis.com/. Older versions of Helm will have the incubator repository configured to a different location.

Run the following in a terminal/shell:

```
helm repo remove incubator && helm repo add incubator https://kubernetes-charts-incubator.storage.googleapis.com/
```

### Installing fabrikate-jaeger

1. In your stack's `component.json`, include `fabrikate-jaeger`:

```json
{
  "name": "my-cool-stack",
  "subcomponents": [
    {
      "name": "fabrikate-jaeger",
      "source": "https://github.com/microsoft/fabrikate-definitions",
      "path": "definitions/fabrikate-jaeger",
      "method": "git"
    }
  ]
}
```

2. In a terminal window, install the stack dependencies:

```
fab install
```

3. In a terminal window, generate the stack:

```
fab generate prod
```

4. Apply the generated stack manifests:

```
kubectl apply -f ./generated/prod/ --recursive
```

### License

MIT
