package test

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/zkbnb-crypto/wasm/legend/legendTxTypes"
)

func TestAddLiquiditySegmentFormat(t *testing.T) {

	var segmentFormat *legendTxTypes.AddLiquiditySegmentFormat
	segmentFormat = &legendTxTypes.AddLiquiditySegmentFormat{
		FromAccountIndex:  0,
		PairIndex:         0,
		AssetAId:          1,
		AssetAAmount:      "10000",
		AssetBId:          2,
		AssetBAmount:      "100",
		LpAmount:          "1000",
		GasAccountIndex:   1,
		GasFeeAssetId:     3,
		GasFeeAssetAmount: "3",
		ExpiredAt:         1654656781000, // milli seconds
		Nonce:             1,
	}

	res, err := json.Marshal(segmentFormat)
	assert.Nil(t, err)

	log.Println(string(res))
}
