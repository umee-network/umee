# Release Notes

Release v0.2.2 of the Umee application. The release includes the following changes:

## Improvements

- Upgrade to Gravity Bridge [v0.2.20](https://github.com/PeggyJV/gravity-bridge/releases/tag/v0.2.20)
  - This release includes a new `gas_limit` field in the `cosmos` section of the
  `gorc` configuration. By default, this value is 500000 which should suffice for
  networks with an average validator set size (100-150). Large validator set sizes
  should increase this value so gravity transactions do not fail due to gas issues.
