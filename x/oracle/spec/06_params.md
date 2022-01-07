
<!--
order: 7
-->

# Parameters

The oracle module contains the following parameters:

| Key                      | Type         | Example                |
|--------------------------|--------------|------------------------|
| VotePeriod               | string (int) | "5"                    |
| VoteThreshold            | string (dec) | "0.500000000000000000" |
| RewardBand               | string (dec) | "0.020000000000000000" |
| RewardDistributionWindow | string (int) | "5256000"              |
| AcceptList               | []DenomList  | [{"base_denom": "uumee", symbol_denom": "UMEE", "exponent": "6"}] |
| SlashFraction            | string (dec) | "0.001000000000000000" |
| SlashWindow              | string (int) | "100800"               |
| MinValidPerWindow        | string (int) | "0.050000000000000000" |
