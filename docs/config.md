# Config Definitions

Configuration files in Fabrikate allow you to define the structure of your
deployment once via [components](./component.md) while enabling elements of it
to vary across different environments like QA, staging, and production or across
on-prem and a public cloud -- or both.

By convention, configuration is placed into a directory called `config` with the
name of the environment that it applies to. Also by convention, if a
`common.yaml` (or `common.json`) config definition exists, it is applied
globally as config.

The schema for these config definitions is fairly simple:

- `config`: A set of configuration values for the component in the parent
  directory that is intended to be used in conjunction with a generator. For
  example, with a Helm generator, these configuration values will be applied
  through a `values.yaml` file to the helm template specified.
- `namespace`: The namespace that should be applied for this component.
- `injectNamespace`: Directs Fabrikate to inject the specified namespace into
  every resource manifest generated for this component. This is intended for
  generators that don't support applying namespaces or where the template for
  the generator doesn't parameterize the namespace such that it is user
  accessible.
- `subcomponents`: A set of key/value pairs for the subcomponents of this
  component that specify the configuration for those components. Each of the
  values of these keys is a config definition in its own right and has the same
  schema as this config definition.

Configuration in Fabrikate is collected from the top of the hierarchy down,
meaning if a config definition lower in the hierarchy specifies a value for a
key of configuration that has been already collected, the configuration provided
higher in the hierarchy wins out. The reasoning behind this is because
configuration higher in the hierarchy has a higher level of context over how the
portions of the deployment definition should work with each other.

Configuration can (and is encouraged to be) factored out into its individual
concerns. For example, to compose a set of resource manifests for a Production
environment in East US in Azure, you might factor your configuration into
`prod`, `east`, and `azure` configuration files such that when you need to build
a Production environment in West US, you can simply swap out the `east` config
for a `west` config.

## Examples

### Jaeger

In this
[config definition](https://github.com/bnookala/fabrikate-jaeger/blob/master/config/common.yaml),
configuration is applied to the subcomponent `jaeger` for
`collector.annotations.sidecar.istio.io/inject` with the value `false` which
will be passed to the helm `values.yaml` file that is applied to this helm
chart. The namespace `jaeger` is also injected into the namespace for every
resource manifest generated as the helm chart templates written for Jaeger were
not parameterized to inject a namespace.

```yaml
config:
subcomponents:
  jaeger:
    namespace: "jaeger"
    injectNamespace: true
    config:
      collector:
        annotations:
          sidecar.istio.io/inject: "false"
```

### Istio

In this
[config definition](https://github.com/evanlouie/fabrikate-istio/blob/master/config/common.yaml)
we are applying the namespace `istio-system` and also applying the config
`global.k8sIngress.enable`: `true` in the `values.yaml` file that is applied to
the helm chart. It also applies the namespace `istio-system` to the subcomponent
`istio-crd`.

```yaml
namespace: istio-system
config:
  name: istio
  global:
    k8sIngress:
      enabled: true # Create the auto-generated ingress
subcomponents:
  istio-crd:
    namespace: istio-system
```

### Disable subcomponents per environment

It is possible to disable subcomponent per environment with a simple `disabled: true` in the
environment config file

```yaml
subcomponents:
  redis:
    disabled: true
```
