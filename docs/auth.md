# Authentication / Personal Access Tokens / `access.yaml`

In order to consume private git repositories, Fabrikate facilitates the use of
environment variables containing [personal access tokens][tokens] in an
`access.yaml` file which lives in the root of your component along side your
`component.yaml`.

If this `access.yaml` file is present in the root component you call
`fab install` on, any components installed (whether in the root component or any
nested subcomponent) which have a matching `source` attribute to any of the keys
found in the `access.yaml` map will use the value contained in the corresponding
env variable as a PAT for authentication.

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

Unless `ENV_VAR_CONTAINING_MY_PAT` and `ANOTHER_ENV_VAR` are set to the correct
tokens in your env, `fab install` will fail. Fabrikate will still attempt to
clone the repos and log warnings that the env vars are missing; however if they
are actually private, it will fail and exit with an authentication error.

Replacing `<token>` with the appropriate values, the following will work:

```bash
ENV_VAR_CONTAINING_MY_PAT=<token> \
ANOTHER_ENV_VAR=<token> \
fab install
```

## Subcomponents specifying `access.yaml`

Any subcomponent in your component tree can also specify an `access.yaml`. This
is important to note when consuming other components which may specify private
repositories. If this occurs, the best way to debug any missing env vars is to
run `fab install` and look at the warning logs to see which components are
asking for which variables and tokens (missing env vars will be logged as
warnings during install).

[tokens]:
  https://help.github.com/en/articles/creating-a-personal-access-token-for-the-command-line
