/*
 * Copyright © 2021 Zecrey Protocol
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package legendTxTypes

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"log"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
)

type RemoveLiquiditySegmentFormat struct {
	FromAccountIndex  int64  `json:"from_account_index"`
	PairIndex         int64  `json:"pair_index"`
	AssetAId          int64  `json:"asset_a_id"`
	AssetAMinAmount   string `json:"asset_a_min_amount"`
	AssetBId          int64  `json:"asset_b_id"`
	AssetBMinAmount   string `json:"asset_b_min_amount"`
	LpAmount          string `json:"lp_amount"`
	AssetAAmountDelta string `json:"asset_a_amount_delta"`
	AssetBAmountDelta string `json:"asset_b_amount_delta"`
	GasAccountIndex   int64  `json:"gas_account_index"`
	GasFeeAssetId     int64  `json:"gas_fee_asset_id"`
	GasFeeAssetAmount string `json:"gas_fee_asset_amount"`
	ExpiredAt         int64  `json:"expired_at"`
	Nonce             int64  `json:"nonce"`
}

func ConstructRemoveLiquidityTxInfo(sk *PrivateKey, segmentStr string) (txInfo *RemoveLiquidityTxInfo, err error) {
	var segmentFormat *RemoveLiquiditySegmentFormat
	err = json.Unmarshal([]byte(segmentStr), &segmentFormat)
	if err != nil {
		log.Println("[ConstructRemoveLiquidityTxInfo] err info:", err)
		return nil, err
	}
	assetAMinAmount, err := StringToBigInt(segmentFormat.AssetAMinAmount)
	if err != nil {
		log.Println("[ConstructBuyNftTxInfo] unable to convert string to big int:", err)
		return nil, err
	}
	assetAMinAmount, _ = CleanPackedAmount(assetAMinAmount)
	assetBMinAmount, err := StringToBigInt(segmentFormat.AssetBMinAmount)
	if err != nil {
		log.Println("[ConstructBuyNftTxInfo] unable to convert string to big int:", err)
		return nil, err
	}
	assetBMinAmount, _ = CleanPackedAmount(assetBMinAmount)
	lpAmount, err := StringToBigInt(segmentFormat.LpAmount)
	if err != nil {
		log.Println("[ConstructBuyNftTxInfo] unable to convert string to big int:", err)
		return nil, err
	}
	lpAmount, _ = CleanPackedAmount(lpAmount)
	assetAAmountDelta, err := StringToBigInt(segmentFormat.AssetAAmountDelta)
	if err != nil {
		log.Println("[ConstructBuyNftTxInfo] unable to convert string to big int:", err)
		return nil, err
	}
	assetAAmountDelta, _ = CleanPackedAmount(assetAAmountDelta)
	assetBAmountDelta, err := StringToBigInt(segmentFormat.AssetBAmountDelta)
	if err != nil {
		log.Println("[ConstructBuyNftTxInfo] unable to convert string to big int:", err)
		return nil, err
	}
	assetBAmountDelta, _ = CleanPackedAmount(assetBAmountDelta)
	gasFeeAmount, err := StringToBigInt(segmentFormat.GasFeeAssetAmount)
	if err != nil {
		log.Println("[ConstructBuyNftTxInfo] unable to convert string to big int:", err)
		return nil, err
	}
	gasFeeAmount, _ = CleanPackedFee(gasFeeAmount)
	txInfo = &RemoveLiquidityTxInfo{
		FromAccountIndex:  segmentFormat.FromAccountIndex,
		PairIndex:         segmentFormat.PairIndex,
		AssetAId:          segmentFormat.AssetAId,
		AssetAMinAmount:   assetAMinAmount,
		AssetBId:          segmentFormat.AssetBId,
		AssetBMinAmount:   assetBMinAmount,
		LpAmount:          lpAmount,
		AssetAAmountDelta: assetAAmountDelta,
		AssetBAmountDelta: assetBAmountDelta,
		GasAccountIndex:   segmentFormat.GasAccountIndex,
		GasFeeAssetId:     segmentFormat.GasFeeAssetId,
		GasFeeAssetAmount: gasFeeAmount,
		Nonce:             segmentFormat.Nonce,
		ExpiredAt:         segmentFormat.ExpiredAt,
		Sig:               nil,
	}
	// compute call data hash
	hFunc := mimc.NewMiMC()
	// compute msg hash
	msgHash, err := txInfo.Hash(hFunc)
	if err != nil {
		log.Println("[ConstructRemoveLiquidityTxInfo] unable to compute hash:", err)
		return nil, err
	}
	// compute signature
	hFunc.Reset()
	sigBytes, err := sk.Sign(msgHash, hFunc)
	if err != nil {
		log.Println("[ConstructRemoveLiquidityTxInfo] unable to sign:", err)
		return nil, err
	}
	txInfo.Sig = sigBytes
	return txInfo, nil
}

type RemoveLiquidityTxInfo struct {
	FromAccountIndex  int64
	PairIndex         int64
	AssetAId          int64
	AssetAMinAmount   *big.Int
	AssetBId          int64
	AssetBMinAmount   *big.Int
	LpAmount          *big.Int
	AssetAAmountDelta *big.Int
	AssetBAmountDelta *big.Int
	KLast             *big.Int
	TreasuryAmount    *big.Int
	GasAccountIndex   int64
	GasFeeAssetId     int64
	GasFeeAssetAmount *big.Int
	ExpiredAt         int64
	Nonce             int64
	Sig               []byte
}

func (txInfo *RemoveLiquidityTxInfo) Validate() error {
	if txInfo.FromAccountIndex < minAccountIndex {
		return fmt.Errorf("FromAccountIndex should not be less than %d", minAccountIndex)
	}
	if txInfo.FromAccountIndex > maxAccountIndex {
		return fmt.Errorf("FromAccountIndex should not be larger than %d", maxAccountIndex)
	}

	if txInfo.PairIndex < minPairIndex {
		return fmt.Errorf("PairIndex should not be less than %d", minPairIndex)
	}
	if txInfo.PairIndex > maxPairIndex {
		return fmt.Errorf("PairIndex should not be larger than %d", maxPairIndex)
	}

	if txInfo.AssetAMinAmount == nil {
		return fmt.Errorf("AssetAMinAmount should not be nil")
	}
	if txInfo.AssetAMinAmount.Cmp(minAssetAmount) < 0 {
		return fmt.Errorf("AssetAMinAmount should not be less than %s", minAssetAmount.String())
	}
	if txInfo.AssetAMinAmount.Cmp(maxAssetAmount) > 0 {
		return fmt.Errorf("AssetAMinAmount should not be larger than %s", maxAssetAmount.String())
	}

	if txInfo.AssetBMinAmount == nil {
		return fmt.Errorf("AssetBMinAmount should not be nil")
	}
	if txInfo.AssetBMinAmount.Cmp(minAssetAmount) < 0 {
		return fmt.Errorf("AssetBMinAmount should not be less than %s", minAssetAmount.String())
	}
	if txInfo.AssetBMinAmount.Cmp(maxAssetAmount) > 0 {
		return fmt.Errorf("AssetBMinAmount should not be larger than %s", maxAssetAmount.String())
	}

	if txInfo.LpAmount == nil {
		return fmt.Errorf("LpAmount should not be nil")
	}
	if txInfo.LpAmount.Cmp(minAssetAmount) < 0 {
		return fmt.Errorf("LpAmount should not be less than %s", minAssetAmount.String())
	}
	if txInfo.LpAmount.Cmp(maxAssetAmount) > 0 {
		return fmt.Errorf("LpAmount should not be larger than %s", maxAssetAmount.String())
	}

	if txInfo.GasAccountIndex < minAccountIndex {
		return fmt.Errorf("GasAccountIndex should not be less than %d", minAccountIndex)
	}
	if txInfo.GasAccountIndex > maxAccountIndex {
		return fmt.Errorf("GasAccountIndex should not be larger than %d", maxAccountIndex)
	}

	if txInfo.GasFeeAssetId < minAssetId {
		return fmt.Errorf("GasFeeAssetId should not be less than %d", minAssetId)
	}
	if txInfo.GasFeeAssetId > maxAssetId {
		return fmt.Errorf("GasFeeAssetId should not be larger than %d", maxAssetId)
	}

	if txInfo.GasFeeAssetAmount == nil {
		return fmt.Errorf("GasFeeAssetAmount should not be nil")
	}
	if txInfo.GasFeeAssetAmount.Cmp(minPackedFeeAmount) < 0 {
		return fmt.Errorf("GasFeeAssetAmount should not be less than %s", minPackedFeeAmount.String())
	}
	if txInfo.GasFeeAssetAmount.Cmp(maxPackedFeeAmount) > 0 {
		return fmt.Errorf("GasFeeAssetAmount should not be larger than %s", maxPackedFeeAmount.String())
	}

	if txInfo.Nonce < minNonce {
		return fmt.Errorf("Nonce should not be less than %d", minNonce)
	}
	return nil
}

func (txInfo *RemoveLiquidityTxInfo) VerifySignature(pubKey string) error {
	// compute hash
	hFunc := mimc.NewMiMC()
	msgHash, err := txInfo.Hash(hFunc)
	if err != nil {
		return err
	}
	// verify signature
	hFunc.Reset()
	pk, err := ParsePublicKey(pubKey)
	if err != nil {
		return err
	}
	isValid, err := pk.Verify(txInfo.Sig, msgHash, hFunc)
	if err != nil {
		return err
	}

	if !isValid {
		return errors.New("invalid signature")
	}
	return nil
}

func (txInfo *RemoveLiquidityTxInfo) GetTxType() int {
	return TxTypeRemoveLiquidity
}

func (txInfo *RemoveLiquidityTxInfo) GetFromAccountIndex() int64 {
	return txInfo.FromAccountIndex
}

func (txInfo *RemoveLiquidityTxInfo) GetNonce() int64 {
	return txInfo.Nonce
}

func (txInfo *RemoveLiquidityTxInfo) GetExpiredAt() int64 {
	return txInfo.ExpiredAt
}

func (txInfo *RemoveLiquidityTxInfo) Hash(hFunc hash.Hash) (msgHash []byte, err error) {
	hFunc.Reset()
	var buf bytes.Buffer
	packedAAmount, err := ToPackedAmount(txInfo.AssetAMinAmount)
	if err != nil {
		log.Println("[ComputeTransferMsgHash] unable to packed amount:", err.Error())
		return nil, err
	}
	packedBAmount, err := ToPackedAmount(txInfo.AssetBMinAmount)
	if err != nil {
		log.Println("[ComputeTransferMsgHash] unable to packed amount:", err.Error())
		return nil, err
	}
	lpAmount, err := ToPackedAmount(txInfo.LpAmount)
	if err != nil {
		log.Println("[ComputeTransferMsgHash] unable to packed amount:", err.Error())
		return nil, err
	}
	packedFee, err := ToPackedFee(txInfo.GasFeeAssetAmount)
	if err != nil {
		log.Println("[ComputeTransferMsgHash] unable to packed amount:", err.Error())
		return nil, err
	}
	WriteInt64IntoBuf(&buf, txInfo.FromAccountIndex)
	WriteInt64IntoBuf(&buf, txInfo.PairIndex)
	WriteInt64IntoBuf(&buf, packedAAmount)
	WriteInt64IntoBuf(&buf, packedBAmount)
	WriteInt64IntoBuf(&buf, lpAmount)
	WriteInt64IntoBuf(&buf, txInfo.GasAccountIndex)
	WriteInt64IntoBuf(&buf, txInfo.GasFeeAssetId)
	WriteInt64IntoBuf(&buf, packedFee)
	WriteInt64IntoBuf(&buf, txInfo.ExpiredAt)
	WriteInt64IntoBuf(&buf, txInfo.Nonce)
	WriteInt64IntoBuf(&buf, ChainId)
	hFunc.Write(buf.Bytes())
	msgHash = hFunc.Sum(nil)
	return msgHash, nil
}
