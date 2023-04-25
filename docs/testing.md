# Umee Testing Guildelines

Rules of thumb:

1. Design packages to make it easy to unit-test.
   - Unit tests should represent majority of the test.
   - Unit tests can be in the same package (to be able to test private methods) → no need to move
     unit tests to `_test` pacakge.
2. Integration tests should be in `_test` package.
3. Reduce tight coupling (will make it ) in the software.
4. Add benchmarks (std lib `BenchmarkXXX`) for all functions we are not sure about the complexity.

Testing libraries:

1. Make sure you are up to date with std lib [testing](https://pkg.go.dev/testing) functionality:
   - Use `FuzzyXXX` tests to generate scenarios
   - Use `t.Skip` to skip long or not stable tests.
     **NOTE**: We should always run all tests before release. So the idea is that long tests won't impact daily CI, but we can make rule that this will be run only once a day and before release.
1. Use [gotest.tools](https://pkg.go.dev/gotest.tools/assert) instead of `stretchr/testify/suite`.
1. Use [gomock](github.com/golang/mock/mockgen) to generate mock interfaces.
1. Use [rapid](https://github.com/flyingmutant/rapid) for fuzzy testing (especially in unit tests).
1. Use module fixtures. Example in [x/leverage/fixtures](../x/leverage/fixtures)

Our CI should use all tests (unit, integration, e2e) except those which are very long or not stable.
Long tests should be used periodically (eg once every second day) instead, to not significatly decrease the pull request experience.

Finally all tests (including QA) must be run on release tag (including `beta` and `rc`).

## Unit Tests

Unit testing is a software testing method by which individual units of source code—sets of one or more computer program modules together with associated control data, usage procedures, and operating procedures—are tested to determine whether they are fit for use.

Unit tests should not have external dependencies (or mock them) except the function or object under the test.

Mock implementations should be provided for all dependencies.

## Integration Tests

Tests which involve at least 2 dependencies tested together. Usually they test integration between 2 or more objects in their expected setup.

Ideally we can test integration between local components without involving the whole app.

At Umee our Keeper tests are usually app level _integration tests_ -- this means that usually we need to setup the whole app to execute the tests, limiting the control of objects under the tests. Our goal should be rather to be able to test modules without setting up an app.

## Simulation Tests

Simulations are very high level App integration tests, where:

- we only run App in it's expected setup (without the normally expected processes, like price-feeder)
- App is producing blocks on the chain until a specified height is reached.
  At each block the sim test runner ([SimulateFromSeed](https://pkg.go.dev/github.com/cosmos/cosmos-sdk/x/simulation#SimulateFromSeed)) sequentially ask each module to generate random transactions (represented as operations) and execute them directly in the app transaction runner.
- during the sim runner, we don't execute any user defined assertions. Instead we expect the chain to function correctly (don't return errors nor panics). Hence, when creating a new operation (random sim transaction), we need to check conditions to satisfy that a transaction won't be rejected (like checking bank balance etc...).

Simulation tests are defined in each module `simulations` package.

## E2E tests

End-to-end test requires the test to setup the system components as they are in production. Real database, services, queues, etc. The chain network (with all required apps, like price-feeder) is setup using [dockertest](https://github.com/ory/dockertest) embedding actual binaries.

These tests execute high level flow by broadcasting transactions to a node port or querying chain directly through the RPC.

E2E tests live in `tests/e2e` package.

As noted in the general guidelines, E2E tests that run for a long time should be skipped in the CI that runs every PR and automatically ran every time a release is tagged, or every night.

## QA tests

For production grade releases we need to have a testable backstage environment with some reliable testing data, accounts and tooling. At the same time we would like programmatic functional tests that run queries and transactions against the backstage environment that can be used for examples presentation, load tests, and smoke tests.

Finally we need to be able to measure the system behavior:

- Errors monitoring
- Regression tests
- Update measures

It’s important to note that this setup could (and should) be packaged and be reusable in all our projects.

### Testing data & tools

As a developer I would like to quickly run the blockchain with some test data, accounts and processes and be able to test new features and replicate some user behaviors (eg for bug exploration / fixing).
Requirements:

- Provide a test data for each module, loadable on latest master. This can be
- Provide tools to dump, store (in some remote service) and restore a blockchain state.
  NOTE: we can (but not necessary must) use the snapshot feature. DB dump should be enough. This process should be fully automated - both for local dev environment as well as for a staging environment.
- Other tools to easily interact with the blockchain (I think this is mostly done through CLI and module “plugin” system).
- test updates and migrations.
