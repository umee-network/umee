
<!--
order: 7
-->

# Parameters

The oracle module contains the following parameters:

| Key                         | Type         | Example                |
|-----------------------------|--------------|------------------------|
| vote_period                 | string (int) | "5"                    |
| vote_threshold              | string (dec) | "0.500000000000000000" |
| reward_band                 | string (dec) | "0.020000000000000000" |
| reward_distribution_ window | string (int) | "5256000"              |
| accept_list                 | []DenomList  | [{"base_denom": "uumee", symbol_denom": "UMEE", "exponent": "6"}] |
| slash_fraction              | string (dec) | "0.001000000000000000" |
| slash_window                | string (int) | "100800"               |
| min_valid_per_window        | string (int) | "0.050000000000000000" |
