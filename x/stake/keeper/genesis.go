package keeper

import (
	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// InitGenesis - store genesis parameters
func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) {
	store := ctx.KVStore(k.storeKey)
	k.SetPool(ctx, data.Pool)
	k.SetNewParams(ctx, data.Params)
	for _, validator := range data.Validators {

		// set validator
		k.SetValidator(ctx, validator)

		// manually set indexes for the first time
		k.SetValidatorByPubKeyIndex(ctx, validator)
		k.SetValidatorByPowerIndex(ctx, validator, data.Pool)
		if validator.Status() == sdk.Bonded {
			store.Set(GetValidatorsBondedIndexKey(validator.PubKey), validator.Owner)
		}
	}
	for _, bond := range data.Bonds {
		k.SetDelegation(ctx, bond)
	}
	k.UpdateBondedValidatorsFull(ctx)
}

// WriteGenesis - output genesis parameters
func (k Keeper) WriteGenesis(ctx sdk.Context) types.GenesisState {
	pool := k.GetPool(ctx)
	params := k.GetParams(ctx)
	validators := k.GetAllValidators(ctx)
	bonds := k.GetAllDelegations(ctx)
	return types.GenesisState{
		pool,
		params,
		validators,
		bonds,
	}
}

// WriteValidators - output current validator set
func (k Keeper) WriteValidators(ctx sdk.Context) (vals []tmtypes.GenesisValidator) {
	k.IterateValidatorsBonded(ctx, func(_ int64, validator sdk.Validator) (stop bool) {
		vals = append(vals, tmtypes.GenesisValidator{
			PubKey: validator.GetPubKey(),
			Power:  validator.GetPower().Evaluate(),
			Name:   validator.GetMoniker(),
		})
		return false
	})
	return
}
