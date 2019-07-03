# Authentication / Personal Access Tokens / `access.yaml`

In order to consume private git repositories, Fabrikate facilitates the use of
environment variables containing [personal access tokens][tokens] in an
`access.yaml` file which lives in the root of your component along side your
`component.yaml`.

## `access.yaml`

The `access.yaml` file is a flat mapping of Component `source` strings to
environment variable.

A sample could look like this:

`my-component/component.yaml`:

```yaml
name: my-private-component
method: git
source: https://github.com/microsoft/my-private-git-repo
subcomponents:
  - name: another-private-component
    method: git
    source: https://github.com/microsoft/another-private-git-repo
  - name: one-more-private-component
    method: git
    source: https://github.com/microsoft/once-more-private-git-repo
```

`my-component/access.yaml`:

```yaml
https://github.com/microsoft/my-private-git-repo: ENV_VAR_CONTAINING_MY_PAT
https://github.com/microsoft/another-private-git-repo: ANOTHER_ENV_VAR
https://github.com/microsoft/once-more-private-git-repo: ANOTHER_ENV_VAR
```

If this `access.yaml` file is present in the root component you call
`fab install` on, any components installed (whether in the root component or any
nested subcomponent) which have a matching `source` attribute to any of the keys
found in the `access.yaml` map will use the value contained in the corresponding
env variable as a PAT for authentication.

[tokens]: https://help.github.com/en/articles/creating-a-personal-access-token-for-the-command-line
