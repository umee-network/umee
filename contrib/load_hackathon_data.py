#!/usr/bin/env python3
import json

genesis_path = "./node-data/umeetest-1/n0/config/genesis.json"

genesis = ""
with open(genesis_path) as f:
    genesis = json.load(f)

registry = ""
with open("./registered_tokens.json") as f:
    registry = json.load(f)

genesis["app_state"]["leverage"]["registry"] = registry


with open(genesis_path, "w") as f:
    json.dump(genesis, f, indent=2)
