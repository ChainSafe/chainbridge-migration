package scripts

import (
	"bridge-scripts/util"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"strconv"
	"strings"
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
				amountOrTokenID, err := strconv.ParseInt(token.AmountOrTokenID, 10, 64)
				if err != nil {
					return err
				}
				txHash, err := util.ExecuteOnBridgeContract(
					chain,
					pk,
					"adminWithdraw",
					common.HexToAddress(token.HandlerAddress),
					common.HexToAddress(token.TokenAddress),
					common.HexToAddress(token.Recipient),
					big.NewInt(amountOrTokenID),
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
