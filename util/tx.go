package util

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"strings"
)

func ExecuteOnBridgeContract(chain RawChainConfig, pk string, method string, args ...interface{}) (*common.Hash, error) {
	bAbi, err := abi.JSON(strings.NewReader(BridgeABI))
	if err != nil {
		return nil, err
	}
	txData, err := bAbi.Pack(method, args...)
	if err != nil {
		return nil, err
	}

	client, err := ethclient.Dial(chain.Endpoint)
	if err != nil {
		return nil, err
	}

	privateKey, err := crypto.HexToECDSA(pk)
	if err != nil {
		return nil, err
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("error casting public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, err
	}

	gasFeeCap, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}
	gasTipCap, err := client.SuggestGasTipCap(context.Background())
	if err != nil {
		return nil, err
	}

	bridgeAddress := chain.Opts["bridge"]
	if bridgeAddress == "" {
		return nil, errors.New("bridge address not defined")
	}
	toAddress := common.HexToAddress(bridgeAddress)

	amount := big.NewInt(0)
	gasLimit := uint64(2100000)

	tx := types.NewTx(&types.DynamicFeeTx{
		Nonce:     nonce,
		To:        &toAddress,
		GasFeeCap: gasFeeCap,
		GasTipCap: gasTipCap,
		Gas:       gasLimit,
		Value:     amount,
		Data:      txData,
	})

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, err
	}

	signedTx, err := types.SignTx(tx, types.NewLondonSigner(chainID), privateKey)
	if err != nil {
		return nil, err
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, err
	}

	hash := signedTx.Hash()
	return &hash, nil
}
