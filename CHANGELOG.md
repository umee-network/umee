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

To release a new version, all the relevant changelog entries from the Unreleased
should be moved into a RELEASE_NOTES.md file and the version should be tagged
and released.

Once a version has been released, the changelog entries from RELEASE_NOTES.md
should be moved into the CHANGELOG.md under the tagged released and removed
from the Unreleased section. The tagged release in the CHANGELOG.md should take
the format:

## [<version>](https://github.com/cosmos/cosmos-sdk/releases/tag/<version>) - YYYY-MM-DD

Ref: https://keepachangelog.com/en/1.0.0/
-->

# Changelog

## [Unreleased]
