# Contributing

Thanks for your interest in contributing! This is an open source project, so we appreciate community contributions.

Pull requests for bug fixes are welcome, but before submitting new features or changes to current functionalities [open an issue](https://github.com/DataDog/KubeHound/issues/new)
and discuss your ideas or propose the changes you wish to make. After a resolution is reached a PR can be submitted for review. PRs created before a decision has been reached may be closed.

For commit messages, try to use the same conventions as most Go projects, for example:

```
contrib/database/sql: use method context on QueryContext and ExecContext

QueryContext and ExecContext were using the wrong context to create
spans. Instead of using the method's argument they were using the
Prepare context, which was wrong.

Fixes #113
```

Please apply the same logic for Pull Requests and Issues: start with the package name, followed by a colon and a description of the change, just like
the official [Go language](https://github.com/golang/go/pulls).

All new code is expected to be covered by tests.

## PR Checks

We expect all PR checks to pass before we merge a PR, which can be investigated by following the `Details` links to CircleCI and CodeCov for unit/integration tests and code coverage checks.

![Screen Shot 2021-08-31 at 10 35 37 AM](https://user-images.githubusercontent.com/1819836/131533266-7c87305d-37df-4bd5-a9ea-6fb8e51e4b50.png)

Please feel free to comment on a PR if there is any difficulty or confusion about any of the checks.

## What to expect

We try to review new PRs within a week of them being opened. If more than two weeks have passed with no reply, please feel free to comment on the PR to bubble it up.

If a PR sits open for more than a month awaiting work or replies by the author, the PR may be closed due to staleness. If you would like to work on it again in the future, feel free to open a new PR and someone will review.

## Adding an Attack

To add a new attack to KubeHound, please do the following:

+ Document the attack in the [edges documentation](./edges/) directory
+ Define the attack constraints in the graph database [schema builder](../deployments/kubehound/janusgraph/kubehound-db-init.groovy)
+ Create an implementation of the [edge.Builder](../pkg/kubehound/graph/edge/builder.go) interface that determines whether attacks are possible by quering the store database and writes any found as edges into the graph database
+ Create the [resources](../test/setup/test-cluster/attacks/) file in the test cluster that will introduce an instance of the attack into the test cluster 
+ Add an [edge system test](../test/system/graph_edge_test.go) that verifies the attack is correctly created by KubeHound
  