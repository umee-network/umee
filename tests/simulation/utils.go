package simulation

import (
	"fmt"

	dbm "github.com/tendermint/tm-db"
)

func printLevelDBStats(db dbm.DB) {
	fmt.Println("\nLevelDB Stats")
	fmt.Println(db.Stats()["leveldb.stats"])
	fmt.Println("LevelDB cached block size", db.Stats()["leveldb.cachedblock"])
}
