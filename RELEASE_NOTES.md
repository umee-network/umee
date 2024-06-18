<!-- markdownlint-disable MD013 -->
<!-- markdownlint-disable MD024 -->
<!-- markdownlint-disable MD040 -->

# Release Notes

The Release Procedure is defined in the [CONTRIBUTING](CONTRIBUTING.md#release-procedure) document.

## v6.5.0

In this release, we are introducing validations for the IBC transfer message receiver address and memo fields. These enhancements aim to address and resolve the recent incident involving spam IBC transfer transactions.

- Maximum length for IBC transfer memo field: 32,768 characters
- Maximum length for IBC transfer receiver address field: 2,048 characters
  