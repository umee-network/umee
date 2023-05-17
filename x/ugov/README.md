# UGov Module

## Abstract

The `x/ugov` is a Cosmos SDK module extending x/gov capabilities.
In particular, this module defines set of parameters controlled by the `x/gov` proposals.
These parameters can be used in various places of the Umee app (other modules or the app core).

## Content

- [Design](#design)
- [Services](#services)
  - [Messages](#messages)
  - [Queries](#queries)
  - [Events](#events)

## Design

### Min Gas Prices

`MsgGovUpdateMinGasPrice` allows to set a consensus controlled transaction min gas prices. All validators
must set their min gas prices that includes `QueryMinGasPrice` and is not smaller than it.
Blocks, that include accepted transactions with smaller gas prices will be out of the consensus.

## Services

### Messages

The RPC [Messages](https://github.com/umee-network/umee/blob/main/proto/umee/ugov/v1/tx.proto) provide an access to the x/gov to change the module parameters.

### Queries

The RPC [Queries](https://github.com/umee-network/umee/blob/main/proto/umee/ugov/v1/query.proto) allow to query module parameters and current outflow sums.

### Events

All events with description are listed in the [events.proto](https://github.com/umee-network/umee/blob/main/proto/umee/ugov/v1/events.proto) file.
