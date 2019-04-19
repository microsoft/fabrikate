# Contibuting to Fabrikate

We do not claim to have all the answers and would gratefully appreciate contributions. This document covers everything you need to know to contribute to Fabrikate.

## Issues and Feature Requests

This project tracks issues exclusively via our project on Github: please [file issues](https://github.com/Microsoft/fabrikate/issues/new/choose) there.

Please do not ask questions via Github issues. Instead, please [join us on Slack](https://publicslack.com/slacks/https-bedrockco-slack-com/invites/new) and ask there.

For issues and feature requests, please follow the general format suggested by the template. Our core team working on Fabrikate utilizes agile development and would appreciate feature requests phrased in the form of a [user story](https://www.mountaingoatsoftware.com/agile/user-stories), as this helps us understand better the context of how the feature would be utilized.

## Pull Requests

Every pull request should be matched with a Github issue. If the pull request is substantial enough to include newly designed elements, this issue should describe the proposed design in enough detail that we can come to an agreement before effort is applied to build the feature. Feel free to start conversations on our Slack #fabrikate channel to get feedback on a design.

When submitting a pull request, please reference the issue the pull request is intended to solve via "Closes #xyz" where is the issue number that is addressed.

## Cloning Fabrikate

Fabrikate is written in [golang](https://golang.org/) so the first step is to make sure you have a fully functioning go development enviroment.

If you intend to make contributions to Fabrikate (versus just build it), the first step is to [fork Fabrikate on Github](https://github.com/Microsoft/fabrikate) into your own account.

Next, clone Fabrikate into your GOPATH (which defaults to $HOME/go) with `go get` (substitute your GitHub username for `Microsoft` below if you forked the repo):

```sh
$ go get github.com/Microsoft/fabrikate
```

If you forked Fabrikate, this will clone your fork into `$GOPATH/<github username>/fabrikate`.  You will want to move to $GOPATH/Microsoft/fabrikate such that the imports in the project work correctly.

### Configuring git
Under `$GOPATH/Microsoft/fabrikate` set up git so that you can push changes to the fork:

```sh
$ git remote add <name> <github_url_of_fork>
```

For example:

```sh
$ git remote add myremote https://github.com/octocat/Spoon-Knife
```

To push changes to the fork:

```sh
$ git push myremote mycurrentbranch
```

## Building Fabrikate

From the root of the project (which if you followed the instructions above should be `$GOPATH/Microsoft/fabrikate`), first fetch project dependencies with:

```sh
$ go get ./...
```

You can then build a Fabrikate executable with:

```sh
$ go build -o fab
```

To build a complete set of release binaries across supported architectures, use our build script, specifying a version number of the release:

```sh
$ scripts/build 0.5.0
```

## Testing Fabrikate

Fabrikate utilizes test driven development to maintain quality across commits. Every code contribution requires covering tests to be accepted by the project and every pull request is built by CI/CD to ensure that the tests pass and that the code is lint free.

You can run project tests by executing the following commands:

```sh
$ go test -v -race ./...
```

And run the linter with:

```sh
$ golangci-lint run
```

## Contributing

This project welcomes contributions and suggestions. Most contributions require you to agree to a Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us the rights to use your contribution. For details, visit https://cla.microsoft.com.

When you submit a pull request, a CLA-bot will automatically determine whether you need to provide a CLA and decorate the PR appropriately (e.g., label, comment). Simply follow the instructions provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.
