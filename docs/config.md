# Config Definitions

Configuration files in Fabrikate allow you to define the structure of your deployment once but still have elements of it vary across environments, like between QA, Staging, and Production or between on-prem and a cloud environment - or both.

By convention, configuration is placed into a directory called `config` with the name of the environment that it applies to.  Also by convention, if a `common.yaml` (or `common.json`) config definition exists, it is always applied as config.

The schema for these config definitions is fairly simple:

* config: A set of configuration values for the current component.
* namespace: The namespace that should be applied for this component.
* injectNamespace: Fabrikate should inject the specified namespace into every resource manifest generated for this component. For generators that don't support applying namespaces or where the template for the generator didn't parameterize the namespace.
* subcomponents: A set of key/value pairs for the subcomponents of this component that specify the configuration for those components. Each of the values of these keys is a config definition in its own right and has the same schema as this config definition.

Configuration in Fabrikate is collected from the top of the hierarchy down, meaning if a config definition lower in the hierarchy specifies a value for an existing key in a component definition, the configuration provided higher in the hierarchy wins out. This is because configuration higher in the hierarchy has a higher level of context over how the portions of the deployment definition should work with each other.

## Examples

### Jaeger

In this [config definition](https://github.com/bnookala/fabrikate-jaeger/blob/master/config/common.yaml), configuration is applied to the subcomponent `jaeger` for `collector.annotations.sidecar.istio.io/inject` with the value `false` which will be passed to the helm `values.yaml` file that is applied to this helm chart.  The namespace `jaeger` is also injected into the namespace for every resource manifest generated as the helm chart templates written for Jaeger were not parameterized to inject a namespace.

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

In this [config definition](https://github.com/evanlouie/fabrikate-istio/blob/master/config/common.yaml) we are applying the namespace `istio-system` and also applying the config `global.k8sIngress.enable`: `true` in the `values.yaml` file that is applied to the helm chart.  It also applies the namespace `istio-system` to the subcomponent `istio-crd`.

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