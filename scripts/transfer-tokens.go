package scripts

import (
	"bridge-scripts/util"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
)

func TransferTokens(v1BridgeConfig *util.V1BridgeConfig, config *util.Config) error {
	if config.Tokens == nil {
		return errors.New("tokens mapping not defined inside configuration")
	}

	for _, chain := range v1BridgeConfig.Chains {
		tokens := config.Tokens[chain.Id]
		if tokens != nil {
			fmt.Printf("Executing token transfer on the chain %s ...\n", chain.Name)
			pk := config.PrivateKeys[chain.Id]
			if pk == "" {
				return errors.New(fmt.Sprintf(
					"Unable to transfer tokens, missing private key for the chain %s\n", chain.Name,
				))
			}
			// execute transfer for all tokens
			for i, token := range tokens {
				fmt.Printf(("token: %s\n"), token)
				// construct main withdrawal data - this is the same for all withdrawals
				var withdrawalData []byte
				withdrawalData = append(withdrawalData, math.PaddedBigBytes(hexutil.MustDecodeBig(token.TokenAddress), 32)...)
				withdrawalData = append(withdrawalData, math.PaddedBigBytes(hexutil.MustDecodeBig(token.Recipient), 32)...)

				if token.Type != "erc1155" {
					amountOrTokenID, err := strconv.ParseInt(token.AmountOrTokenID[0], 10, 64)
					if err != nil {
						return err
					}
					withdrawalData = append(withdrawalData, math.PaddedBigBytes(big.NewInt(amountOrTokenID), 32)...)
				} else {
					if len(token.AmountOrTokenID) != len(token.ERC1155Amounts) {
						return errors.New(fmt.Sprintf("ERC1155 TokenIDs and token amounts arrays not the same lenghts"))
					}
					/* ERC1155 data structure
					8 byte - func sig
					32 byte - offset to data start
					32 byte - data length
					32 byte - offset for first argument start
					32 byte - offset for second argument start
					32 byte - offset for n-th argument start
					32 byte - count of first argument elements
					32 byte - encoding of first argument elements
					32 byte - count of second argument elements
					32 byte - encoding of second argument elements
					32 byte - count of n-th argument elements
					32 byte - encoding of n-th argument elements
					*/
					// offset for arguments start - it starts at 5th slot of 32 bytes
					fistSlotOffset := 5 * 32
					withdrawalData = append(withdrawalData, math.PaddedBigBytes(big.NewInt(int64(fistSlotOffset)), 32)...)
					// offset for second arguments array start - it starts at firstSlotOffset + len(tokenID) * 32 bytes
					secondSlotOffset := fistSlotOffset + 32*(1+len(token.AmountOrTokenID))
					withdrawalData = append(withdrawalData, math.PaddedBigBytes(big.NewInt(int64(secondSlotOffset)), 32)...)
					// offset for second arguments array start - it starts at secondSlotOffset + len(tokenAmount) * 32 bytes
					thirdSlotOffset := secondSlotOffset + 32*(1+len(token.ERC1155Amounts))
					withdrawalData = append(withdrawalData, math.PaddedBigBytes(big.NewInt(int64(thirdSlotOffset)), 32)...)

					// length of tokens IDs
					withdrawalData = append(withdrawalData, math.PaddedBigBytes(big.NewInt(int64(len(token.AmountOrTokenID))), 32)...)
					// encode token IDs
					for _, tokenID := range token.AmountOrTokenID {
						amountOrTokenID, err := strconv.ParseInt(tokenID, 10, 64)
						if err != nil {
							return err
						}
						withdrawalData = append(withdrawalData, math.PaddedBigBytes(big.NewInt(amountOrTokenID), 32)...)
					}

					// length of token amounts
					withdrawalData = append(withdrawalData, math.PaddedBigBytes(big.NewInt(int64(len(token.ERC1155Amounts))), 32)...)
					// encode token IDs
					for _, tokenAmount := range token.ERC1155Amounts {
						tokenAmount, err := strconv.ParseInt(tokenAmount, 10, 64)
						if err != nil {
							return err
						}
						// encode token amounts
						withdrawalData = append(withdrawalData, math.PaddedBigBytes(big.NewInt(tokenAmount), 32)...)
					}
					// ERC1155 additional transfer data - check ERC1155 specific token implementation if it's empty or not
					withdrawalData = append(withdrawalData, math.PaddedBigBytes(big.NewInt(int64(0)), 32)...)
				}

				txHash, err := util.ExecuteOnBridgeContract(
					chain,
					pk,
					"adminWithdraw",
					common.HexToAddress(token.HandlerAddress),
					withdrawalData,
				)
				if err != nil {
					fmt.Printf("[%d] Unable to transfer %s tokens %s to %s\n\tOn the chain %s, because: %v\n",
						i, token.AmountOrTokenID, token.TokenAddress, token.Recipient, chain.Name, err)
				} else {
					fmt.Printf("[%d] Transfer of %s token %s\n"+
						"\tAmount/TokenID: %s\n"+
						"\tTo: %s\n"+
						"\tSubmitted with hash %s on the chain %s\n",
						i, strings.ToUpper(token.Type), token.TokenAddress, token.AmountOrTokenID, token.Recipient, txHash.Hex(), chain.Name)
				}
			}
		} else {
			fmt.Printf("No token transfers defined for chain %s\n", chain.Name)
		}
		util.DisplayLine()
	}
	return nil
}
