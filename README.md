# marina

NOTE: The description here is aspirational, under active development, and an initial release is not yet available. In the meantime, we welcome your feedback and pull requests to help shape how this tool evolves.

Marina makes devops for cloud native applications on Kubernetes easier. It enables writing [DRY](https://en.wikipedia.org/wiki/Don%27t_repeat_yourself) resource definitions and configuration, capturing common resource definitions into abstracted and shareable components, and building a [GitOps](https://www.weave.works/blog/gitops-operations-by-pull-request) deployment workflow that both simplifies operations and makes deployments more auditable.

In particular, Marina simplifies the frontend of the GitOps workflow: it takes a high level description of your deployment, a target environment (eg. `dev` or `prod`), and renders the resource definitions for that deployment. It is intended to run as part of your CI/CD pipeline such that with every commit to your Marina deployment project triggers the generation of Kubernetes resource descriptions that is then reconciled with the current state of your Kubernetes cluster using a tool like [Flux](https://github.com/weaveworks/flux).

The best way to understand what Marina provides is see it in practice, so let's walk through the creation and use of a simple workload.

## Developer Experience

First off, install the `marina` cli on your local machine.  This CLI tool, `docker`, and `git` are the only tools you need to have installed: any other dependencies will be fetched via `docker` images and/or `git`.

```
$ curl -sL https://run.marina.io/install | sh
```

Let's next create our first project, in this case for a hypothetical microservices workload that we want to operationalize.

```
$ cd ~/dev (or wherever you like to keep your development projects)
$ marina scaffold component microservices-workload
$ cd microservces-workload
```

This creates a folder called `microservices-workload`, a `config` folder that holds configuration data for different environments using this definition, and a `component.json` file in the current directory that looks like this:

```
{
    "name": "microservices-workload",
    "version": "1.0.0",
    "components": []
}
```

A component in Marina houses the definition for building the Kubernetes resource definitions for its directory tree scope. 

Let's say we are setting up a cluster running a number of microservices, so the first thing we'd like to define is to add a component to manage the common infrastructure pieces that makes all of our infrastructure more observable.

We want to house all of these in a separate subcomponent called `infra`, so in the same way that we created the top level component of this project, let's scaffold that as well:

```
$ marina scaffold component infra
```

This updates the top level project to include a component:

```
{
     "name": "microservices-workload",
     "version": "1.0.0",
     "components": [{
         "name": "infra",
         "source": "./infra"
     }]
}
```

and creates a subcomponent called `infra` with a similarly scaffolded out `component.json`:

```
$ cd infra
$ cat component.json
{
    "name": "infra",
    "version": "1.0.0",
    "components": []
}
```

Next, let's add an Elasticsearch, Fluentd, and Kibana based logging system as base logging infrastructure for our cluster. We can do that with:

```
$ marina add efk https://github.com/Microsoft/marina-elasticsearch-fluentd-kibana
```

This updates the `component.json` for our `infra` project to include this component:

```
{
    "name": "infra",
    "type": "component",
    "components": [{
        "name": "efk",
        "source": "https://github.com/Microsoft/marina-elasticsearch-fluentd-kibana",
        "tag": "^1.4.1"
    }]
}
```

This looks similar to our own infra subcomponent but instead of the source location being local, it is instead a git endpoint that this component should be fetched from. This enables us to share components between projects, but also do semver locking of them to a specific version to limit the risk of them from evolving underneath us and breaking our deployments, while still picking up small improvements.

This component also takes a number of inputs. Marina has automatically scaffolded these configuration properties into `config/common.json` when we added the component to the project. In a config folder, `common.json` are the defaults that will be applied for all environments in the absence of an overriding definition for the specific environment you are building.

```
{
    "config": {
        "efk": {
            "elasticsearch": {
                "namespace": "elasticsearch",
                "master-storage-class": "default",
                "master-storage-size": "4Gi",
                "data-storage-class": "default",
                "data-storage-size": "4Gi"
            },
            "fluentd": {
                "namespace": "fluentd"
            },
            "kibana": {
                "namespace": "kibana"
            }
        }
    }
}
```

Using shared components like this also provides a level of abstraction away from the implementation.  This component's subcomponents use helm to template out the resource descriptions but we don't have to be concerned with that: we can just focus on providing reasonable values for the inputs to the component and Marina will handle the rest under the hood.

We know that these default sizes and storage classes are not large or fast enough for our `prod` cluster and we want to override them. Let's scaffold out a new `prod` config to fix this:

```
$ marina scaffold config prod
```

This scaffolds out `config/prod.json` for us that has sections for all of our subcomponents (and their subcomponents):

```
{
    "config": {
        "efk": {
            "elasticsearch": {},
            "fluentd": {},
            "kibana": {}
        }
    }
}
```

Let's update the elasticsearch section to have a more `prod` like configuration:

```
{
    "config": {
        "efk": {
            "elasticsearch": {
                "master-storage-class": "managed-premium",
                "master-storage-size": "16Gi",
                "data-storage-class": "managed-premium",
                "data-storage-size": "64Gi"
            },
            "fluentd": {},
            "kibana": {}
        }
    }
}
```

When the `prod` environment is rendered for our project, any values in `prod.json` will override the values in `common.json`. 

Moving back to the root of our project, let's scaffold out a component to deploy all our microservices.

```
$ marina scaffold component services --type "static"
```

Like the `infra` component we created earlier, this is an umbrella component for all our microservices for all the common elements of our services.  It has a type of `static`, which is a subclass of `component` that includes static file based resource definitions that are specified by the `path` property.  It also defines a hook that is run after this and every child component, which we leverage to introduce a Linkerd filter to inject a service mesh sidecar into all of our service deployments.

```
{
    "name": "services",
    "type": "static",
    "path": "./resources",
    "version": "1.0.0",
    "components": [{
        "name": "simple-service",
        "source": "./simple-service"
    }],
    "hooks": [{
        "after": [{
            "name": "linkerd",
            "source": "https://github.com/Microsoft/marina-linkerd-filter"
        }]
    }]
}
```

With our umbrella component defined, let's scaffold out a concrete microservice.

```
$ marina scaffold component simple-service --type=helm
```

This service is templated with helm charts, so instead of choosing the generic component type, we instead specify a more specific `helm` type.  Like the `static` type above, the `helm` type is a subclass of `component` and has `helm` specific properties like the location of the chart to use to deploy the application.

```
{
    "name": "simple-service",
    "type": "helm",
    "chart": "./chart",
    "version": "1.0.0"
}
```

 A `helm` component knows how to take a chart and apply the values specified in the environment's config and materialize the resource definitions from them. All of the config traits we used as we were defining our logging infrastructure apply to this as well.  For example, if we define `config/common.json` as:

```
{
    "config": {
        "simple-service": {
            "serviceName": "simple-service",
            ...
        }
    }
}
```

and a `config/dev.json` as:

```
{
    "config": {
        "simple-service": {
            "replicas": 1,
            "imageName": "timfpark/simple-service:edge"
            ...
        }
    }
}
```

and a `config/prod.json` that looks like:

```
{
    "config": {
        "simple-service": {
            "replicas": 6,
            "imageName": "timfpark/simple-service:stable"
            ...
        }
    }
}
```

Then the rendered resource definitions will have `prod` deployments with 6 replicas running the `stable` tagged image, the `dev` environment will only 1 replica running the `edge` tagged image, but both will have a service name of `simple-service`.

With this - the grand finale (and the real purpose for all of this definition) -- generating all the resource descriptions for all of the components we defined.  You can now do that by simply going to the root of our project and executing:

```
$ marina generate dev
```

In this invocation we are generating all of the resource definitions with a `dev` config.  By default, all of the resulting resource definitions will be placed in a directory called `generated/dev` at the root of the project and organized with the same directory structure as the project itself.

As we mentioned at the beginning, Marina is intended to be executed as part of a CI/CD pipeline with downstream tooling to manage reconciling the resource definitions in our cluster, but if you have a Kubernetes cluster handy, you can also apply the generated resource descriptions directly with:

```
$ cd generated/dev
$ kubectl apply --recursive -f .
```

##  Contributing

This project welcomes contributions and suggestions.  Most contributions require you to agree to a Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us the rights to use your contribution. For details, visit https://cla.microsoft.com.

When you submit a pull request, a CLA-bot will automatically determine whether you need to provide a CLA and decorate the PR appropriately (e.g., label, comment). Simply follow the instructions provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.


