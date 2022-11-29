package keeper

import (
	"encoding/json"
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/trie"
)

// BeginBlock sets the sdk Context and EIP155 chain id to the Keeper.
func (k *Keeper) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	k.WithChainID(ctx)
}

// EndBlock also retrieves the bloom filter value from the transient store and commits it to the
// KVStore. The EVM end block logic doesn't update the validator set, thus it returns
// an empty slice.
func (k *Keeper) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
	// Gas costs are handled within msg handler so costs should be ignored
	infCtx := ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

	bloom := ethtypes.BytesToBloom(k.GetBlockBloomTransient(infCtx).Bytes())
	k.EmitBlockBloomEvent(infCtx, bloom)

	ethTxs := k.GetTxs(ctx)
	txRoot := ethtypes.EmptyRootHash
	if len(ethTxs) != 0 {
		hasher := trie.NewStackTrie(nil)
		txRoot = ethtypes.DeriveSha(ethtypes.Transactions(ethTxs), hasher)	
	}
	k.EmitTxRootEvent(ctx, txRoot)

	ethReceipts := k.GetReceipts(ctx)
	JSONReceipts, err := json.MarshalIndent(ethReceipts, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("End Block Receipts: %s\n", string(JSONReceipts))
	receiptHash := ethtypes.EmptyRootHash
	if len(ethReceipts) != 0 {
		hasher := trie.NewStackTrie(nil)
		receiptHash = ethtypes.DeriveSha(ethtypes.Receipts(ethReceipts), hasher)
	}
	fmt.Printf("End Block receiptHash: %s\n", receiptHash)
	k.EmitReceiptHashEvent(ctx, receiptHash)

	gasLimit := ctx.BlockGasMeter().Limit()
	k.EmitGasLimitEvent(ctx, gasLimit)

	return []abci.ValidatorUpdate{}
}
