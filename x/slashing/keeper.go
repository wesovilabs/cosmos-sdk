package slashing

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/params"
	crypto "github.com/tendermint/go-crypto"
)

// Keeper of the slashing store
type Keeper struct {
	storeKey     sdk.StoreKey
	cdc          *wire.Codec
	validatorSet sdk.ValidatorSet
	params       params.Getter

	// codespace
	codespace sdk.CodespaceType
}

// NewKeeper creates a slashing keeper
func NewKeeper(cdc *wire.Codec, key sdk.StoreKey, vs sdk.ValidatorSet, params params.Getter, codespace sdk.CodespaceType) Keeper {
	keeper := Keeper{
		storeKey:     key,
		cdc:          cdc,
		validatorSet: vs,
		params:       params,
		codespace:    codespace,
	}
	return keeper
}

// handle a validator signing two blocks at the same height
func (k Keeper) handleDoubleSign(ctx sdk.Context, height int64, timestamp int64, pubkey crypto.PubKey) {
	logger := ctx.Logger().With("module", "x/slashing")
	age := ctx.BlockHeader().Time - timestamp

	// Double sign too old
	maxEvidenceAge := k.MaxEvidenceAge(ctx)
	if age > maxEvidenceAge {
		logger.Info(fmt.Sprintf("Ignored double sign from %s at height %d, age of %d past max age of %d", pubkey.Address(), height, age, maxEvidenceAge))
		return
	}

	// Double sign confirmed
	logger.Info(fmt.Sprintf("Confirmed double sign from %s at height %d, age of %d less than max age of %d", pubkey.Address(), height, age, maxEvidenceAge))
	k.validatorSet.Slash(ctx, pubkey, height, k.SlashFractionDoubleSign(ctx))
}

// handle a validator signature, must be called once per validator per block
func (k Keeper) handleValidatorSignature(ctx sdk.Context, pubkey crypto.PubKey, signed bool) {
	logger := ctx.Logger().With("module", "x/slashing")
	height := ctx.BlockHeight()
	if !signed {
		logger.Info(fmt.Sprintf("Absent validator %s at height %d", pubkey.Address(), height))
	}
	address := pubkey.Address()

	// Local index, so counts blocks validator *should* have signed
	// Will use the 0-value default signing info if not present, except for start height
	signInfo, found := k.getValidatorSigningInfo(ctx, address)
	if !found {
		// If this validator has never been seen before, construct a new SigningInfo with the correct start height
		signInfo = NewValidatorSigningInfo(height, 0, 0, 0)
	}
	index := signInfo.IndexOffset % k.SignedBlocksWindow(ctx)
	signInfo.IndexOffset++

	// Update signed block bit array & counter
	// This counter just tracks the sum of the bit array
	// That way we avoid needing to read/write the whole array each time
	previous := k.getValidatorSigningBitArray(ctx, address, index)
	if previous == signed {
		// Array value at this index has not changed, no need to update counter
	} else if previous && !signed {
		// Array value has changed from signed to unsigned, decrement counter
		k.setValidatorSigningBitArray(ctx, address, index, false)
		signInfo.SignedBlocksCounter--
	} else if !previous && signed {
		// Array value has changed from unsigned to signed, increment counter
		k.setValidatorSigningBitArray(ctx, address, index, true)
		signInfo.SignedBlocksCounter++
	}

	minHeight := signInfo.StartHeight + k.SignedBlocksWindow(ctx)
	minSignedPerWindow := k.MinSignedPerWindow(ctx)
	if height > minHeight && signInfo.SignedBlocksCounter < minSignedPerWindow {
		// Downtime confirmed, slash, revoke, and jail the validator
		logger.Info(fmt.Sprintf("Validator %s past min height of %d and below signed blocks threshold of %d", pubkey.Address(), minHeight, minSignedPerWindow))
		k.validatorSet.Slash(ctx, pubkey, height, k.SlashFractionDowntime(ctx))
		k.validatorSet.Revoke(ctx, pubkey)
		signInfo.JailedUntil = ctx.BlockHeader().Time + k.DowntimeUnbondDuration(ctx)
	}

	// Set the updated signing info
	k.setValidatorSigningInfo(ctx, address, signInfo)
}
