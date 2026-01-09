package types

// GetPoolUidString returns a poolUid from pool
func GetPoolUidString(pool Pool) string {
	return pool.Coins[0].Denom + PoolUidSeparator + pool.Coins[1].Denom
}
