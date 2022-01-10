
<!--
order: 7
-->

# Parameters

The oracle module contains the following parameters:

| Key                      | Type             | Example                |
|--------------------------|------------------|------------------------|
| VotePeriod               | string (uint64)  | "5"                    |
| VoteThreshold            | string (sdk.Dec) | "0.500000000000000000" |
| RewardBand               | string (sdk.Dec) | "0.020000000000000000" |
| RewardDistributionWindow | string (uint64)  | "5256000"              |
| AcceptList               | []DenomList      | [{"base_denom": "uumee", symbol_denom": "UMEE", "exponent": "6"}] |
| SlashFraction            | string (sdk.Dec) | "0.001000000000000000" |
| SlashWindow              | string (uint64)  | "100800"               |
| MinValidPerWindow        | string (uint64)  | "0.050000000000000000" |
