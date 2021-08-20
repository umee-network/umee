<!--
Changelog Guiding Principles:

Changelogs are for humans, not machines.
There should be an entry for every single version.
The same types of changes should be grouped.
Versions and sections should be linkable.
The latest version comes first.
The release date of each version is displayed.
Mention whether you follow Semantic Versioning.

Usage:

Change log entries are to be added to the Unreleased section under the
appropriate stanza (see below). Each entry should ideally include a tag and
the Github PR referenced in the following format:

* (<tag>) [#<PR-number>](https://github.com/cosmos/cosmos-sdk/pull/<PR-number>) <changelog entry>

Types of changes (Stanzas):

Features: for new features.
Improvements: for changes in existing functionality.
Deprecated: for soon-to-be removed features.
Bug Fixes: for any bug fixes.
Client Breaking: for breaking Protobuf, CLI, gRPC and REST routes used by clients.
API Breaking: for breaking exported Go APIs used by developers.
State Machine Breaking: for any changes that result in a divergent application state.

To release a new version, ensure an appropriate release branch exists. Add a
release version and date to the existing Unreleased section which takes the form
of:

## [<version>](https://github.com/cosmos/cosmos-sdk/releases/tag/<version>) - YYYY-MM-DD

Once the version is tagged and released, a PR should be made against the main
branch to incorporate the new changelog updates.

Ref: https://keepachangelog.com/en/1.0.0/
-->

# Changelog

## [Unreleased]
