package util

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"strconv"
	"strings"
)

type PendingProposal struct {
	EventName   string
	TxHash      string
	BlockNumber uint64
	Event       ProposalVote
}

type ProposalVote struct {
	OriginChainID  uint8
	DepositNonce   uint64
	ProposalStatus uint8
	ResourceID     [32]byte
	DataHash       [32]byte
}

var ProposalStatusMap = map[uint8]string{
	0: "Inactive",
	1: "Active",
	2: "Passed",
	3: "Executed",
	4: "Cancelled",
}

func DisplayProposals(deposits []PendingProposal) {
	fmt.Printf("%d pending deposits:\n", len(deposits))
	DisplayLine()
	for i, d := range deposits {
		fmt.Printf(
			"[%d] Status: %s OriginChainID: %d DepositNonce: %d ResourceID: %s DataHash: %s \n"+
				"    => Event: %s BlockNumber: %d TxHash: %s\n",
			i,
			ProposalStatusMap[d.Event.ProposalStatus],
			d.Event.OriginChainID,
			d.Event.DepositNonce,
			hexutil.Encode(d.Event.ResourceID[:]),
			hexutil.Encode(d.Event.DataHash[:]),
			d.EventName,
			d.BlockNumber,
			d.TxHash,
		)
	}
}

func DisplayLine() {
	fmt.Println("-----------------------------------------------------------")
}

func Hex2uint64(hexStr string) uint64 {
	// remove 0x suffix if found in the input string
	cleaned := strings.Replace(hexStr, "0x", "", -1)

	// base 16 for hexadecimal
	result, _ := strconv.ParseUint(cleaned, 16, 64)
	return uint64(result)
}

func Hex2uint8(hexStr string) uint8 {
	// remove 0x suffix if found in the input string
	cleaned := strings.Replace(hexStr, "0x", "", -1)

	// base 16 for hexadecimal
	result, _ := strconv.ParseUint(cleaned, 16, 8)
	return uint8(result)
}
