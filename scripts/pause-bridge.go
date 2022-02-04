package scripts

import (
	"bridge-scripts/util"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"strconv"
	"strings"
	"time"
)

func PauseBridge(v1BridgeConfig *util.V1BridgeConfig, config *util.Config) error {
	var hasChainPendingProposals = map[string]bool{}
	for _, c := range v1BridgeConfig.Chains {
		hasChainPendingProposals[c.Id] = true
	}

	for true {
		for _, chain := range v1BridgeConfig.Chains {
			fmt.Printf("Checking for pending proposals on chain %s ...\n", chain.Name)
			client, err := ethclient.Dial(chain.Endpoint)

			startingBlock := config.StartingBlocks[chain.Id]
			if startingBlock == "" {
				startingBlock = "0"
			}
			fromBlock, err := strconv.ParseInt(startingBlock, 10, 64)
			if err != nil {
				return fmt.Errorf(
					"unable to parse starting block for chain %s, because: %v", chain.Id, err,
				)
			}

			fmt.Printf("Querying for proposals from block: %d\n", fromBlock)
			pendingProposals, err := getAllPendingProposals(client, chain, fromBlock)
			if err != nil {
				return err
			}
			hasChainPendingProposals[chain.Id] = len(pendingProposals) != 0
			util.DisplayProposals(pendingProposals)
		}

		anyChainHasPending := false
		for _, hasPending := range hasChainPendingProposals {
			if hasPending {
				anyChainHasPending = true
			}
		}

		util.DisplayLine()
		if anyChainHasPending {
			fmt.Printf("Waiting for %d seconds....\n", 60)
			time.Sleep(60 * time.Second)
			continue
		} else {
			break
		}
	}

	fmt.Println("All proposals have been resolved!")
	util.DisplayLine()

	if config.AutoPauseBridge {
		// pause bridge contracts on all chains
		for _, chain := range v1BridgeConfig.Chains {
			pk := config.PrivateKeys[chain.Id]
			if pk == "" {
				fmt.Printf("Unable to pause bridge contract, missing private key for chain %s\n", chain.Name)
			}
			txHash, err := util.ExecuteOnBridgeContract(chain, pk, "adminPauseTransfers")
			if err != nil {
				fmt.Printf("Unable to pause bridge contract for chain %s, because: %v\n", chain.Name, err)
			} else {
				fmt.Printf("Transaction for pausing bridge contract on chain %s submitted with hash %s\n",
					chain.Name, txHash.Hex())
			}
		}
	}

	return nil
}

func getAllPendingProposals(
	client *ethclient.Client,
	config util.RawChainConfig,
	fromBlock int64,
) ([]util.PendingProposal, error) {

	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(fromBlock),
		ToBlock:   nil,
		Addresses: []common.Address{
			common.HexToAddress(config.Opts["bridge"]),
		},
	}

	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		return nil, err
	}

	bAbi, err := abi.JSON(strings.NewReader(util.BridgeABI))
	if err != nil {
		return nil, err
	}

	proposalData := map[uint8]map[[32]byte]map[uint64]*util.PendingProposal{}
	for _, vLog := range logs {
		eventByID, err := bAbi.EventByID(vLog.Topics[0])
		if err != nil {
			continue
		} else {
			if eventByID.Name == "ProposalEvent" {
				inputs, err := eventByID.Inputs.Unpack(vLog.Data)
				if err != nil {
					return nil, err
				}

				originChainID := util.Hex2uint8(vLog.Topics[1].Hex())
				depositNonce := util.Hex2uint64(vLog.Topics[2].Hex())
				proposalStatus := util.Hex2uint8(vLog.Topics[3].Hex())

				var resourceId [32]byte
				rawResourceId, ok := inputs[0].([32]uint8)
				if !ok {
					return nil, fmt.Errorf("unable to convert resource id")
				}
				copy(resourceId[:], rawResourceId[:])

				var dataHash [32]byte
				rawDataHash, ok := inputs[1].([32]uint8)
				if !ok {
					return nil, fmt.Errorf("unable to convert data hash")
				}
				copy(dataHash[:], rawDataHash[:])

				if proposalData[originChainID] == nil {
					proposalData[originChainID] = map[[32]byte]map[uint64]*util.PendingProposal{}
				}

				if proposalData[originChainID][resourceId] == nil {
					proposalData[originChainID][resourceId] = map[uint64]*util.PendingProposal{}
				}

				if proposalStatus == 1 || proposalStatus == 2 { // Proposal Active or Passed
					proposalData[originChainID][resourceId][depositNonce] = &util.PendingProposal{
						EventName:   eventByID.Name,
						TxHash:      vLog.TxHash.Hex(),
						BlockNumber: vLog.BlockNumber,
						Event: util.ProposalVote{
							OriginChainID:  originChainID,
							DepositNonce:   depositNonce,
							ProposalStatus: proposalStatus,
							ResourceID:     resourceId,
							DataHash:       dataHash,
						},
					}
				} else if proposalStatus == 3 || proposalStatus == 4 { // Proposal Executed or Cancelled
					delete(proposalData[originChainID][resourceId], depositNonce)
				}

			}
		}
	}

	// flatten out map to array of pending proposals
	var pendingProposals []util.PendingProposal
	for _, v := range proposalData {
		for _, m := range v {
			for _, proposal := range m {
				pendingProposals = append(pendingProposals, *proposal)
			}
		}
	}

	return pendingProposals, err
}
