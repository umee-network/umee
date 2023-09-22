package keeper

func (k Keeper) Bond() error {
	if k.ctx.BlockTime().Before(k.getNextBondingTime()) {
		return nil
	}

	//TODO: confirm bonding period = zero
	//      get ongoing campaigns
	//      filter by denom
	//      if find some: collateralize and bond

	return nil
}
