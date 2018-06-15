package slashing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MaxEvidenceAge - Max age for evidence - default 21 days (3 weeks)
func (k Keeper) MaxEvidenceAge(ctx sdk.Context) int64 {
	return k.params.GetInt64WithDefault(ctx, "MaxEvidenceAge", defaultMaxEvidenceAge)
}

// SignedBlocksWindow - sliding window for downtime slashing
func (k Keeper) SignedBlocksWindow(ctx sdk.Context) int64 {
	return k.params.GetInt64WithDefault(ctx, "SignedBlocksWindow", defaultSignedBlocksWindow)
}

// Downtime slashing thershold - default 50%
func (k Keeper) MinSignedPerWindow(ctx sdk.Context) int64 {
	minSignedPerWindow := k.params.GetRatWithDefault(ctx, "MinSignedPerWindow", defaultMinSignedPerWindow)
	signedBlocksWindow := k.SignedBlocksWindow(ctx)
	return sdk.NewRat(signedBlocksWindow).Mul(minSignedPerWindow).Evaluate()
}

// Downtime unbond duration
func (k Keeper) DowntimeUnbondDuration(ctx sdk.Context) int64 {
	return k.params.GetInt64WithDefault(ctx, "DowntimeUnbondDuration", defaultDowntimeUnbondDuration)
}

// SlashFractionDoubleSign - currently default 5%
func (k Keeper) SlashFractionDoubleSign(ctx sdk.Context) sdk.Rat {
	return k.params.GetRatWithDefault(ctx, "SlashFractionDoubleSign", defaultSlashFractionDoubleSign)
}

// SlashFractionDowntime - currently default 1%
func (k Keeper) SlashFractionDowntime(ctx sdk.Context) sdk.Rat {
	return k.params.GetRatWithDefault(ctx, "SlashFractionDowntime", defaultSlashFractionDowntime)
}

const (

	// defaultMaxEvidenceAge = 60 * 60 * 24 * 7 * 3
	// TODO Temporarily set to 2 minutes for testnets.
	defaultMaxEvidenceAge int64 = 60 * 2

	// TODO Temporarily set to 100 blocks for testnets
	defaultSignedBlocksWindow int64 = 100

	// TODO Temporarily set to 10 minutes for testnets
	defaultDowntimeUnbondDuration int64 = 60 * 10
)

var (
	defaultMinSignedPerWindow sdk.Rat = sdk.NewRat(1, 2)

	defaultSlashFractionDoubleSign = sdk.NewRat(1).Quo(sdk.NewRat(20))

	defaultSlashFractionDowntime = sdk.NewRat(1).Quo(sdk.NewRat(100))
)
