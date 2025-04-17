package keeper

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/pelldelegationmanager.sol"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

func TestOperatorHandler_HandleEvent(t *testing.T) {
	contractAddr := ethcommon.HexToAddress("0xCCB019e26DAc3f5AD6b67526d36b27d82C6102A2")
	logs := make([]ethtypes.Log, 0)
	testlog := `[{"address":"0xccb019e26dac3f5ad6b67526d36b27d82c6102a2","blockHash":"0x2444999fc774c1508fff609459ad771fbd98017354c3fc42a58703dd746d888b","blockNumber":"0x37","data":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000","logIndex":"0x0","removed":false,"topics":["0x9370d03062ed1495767fd7109ad3958578bab7d729f5f7d9e204d61bd2afd3fc","0x0000000000000000000000007c4055da5ad452d6901c94d65cb7dbc5ac7bbbaa"],"transactionHash":"0x8d121def15c5e15d47fde151e0feb8a418cbe73dd63a164289f13e1fac156c2d","transactionIndex":"0x0"},{"address":"0xccb019e26dac3f5ad6b67526d36b27d82c6102a2","blockHash":"0x2444999fc774c1508fff609459ad771fbd98017354c3fc42a58703dd746d888b","blockNumber":"0x37","data":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000","logIndex":"0x1","removed":false,"topics":["0xfc64399f88afa882a6960f089803d4c363341aa85ade7826360afc812106ae46","0x0000000000000000000000007c4055da5ad452d6901c94d65cb7dbc5ac7bbbaa"],"transactionHash":"0x8d121def15c5e15d47fde151e0feb8a418cbe73dd63a164289f13e1fac156c2d","transactionIndex":"0x0"},{"address":"0xccb019e26dac3f5ad6b67526d36b27d82c6102a2","blockHash":"0x2444999fc774c1508fff609459ad771fbd98017354c3fc42a58703dd746d888b","blockNumber":"0x37","data":"0x0000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000005268747470733a2f2f7261772e67697468756275736572636f6e74656e742e636f6d2f6d617474686577373235312f4d657461646174612f6d61696e2f466f7572556e69745f4d657461646174612e6a736f6e0000000000000000000000000000","logIndex":"0x2","removed":false,"topics":["0x02a919ed0e2acad1dd90f17ef2fa4ae5462ee1339170034a8531cca4b6708090","0x0000000000000000000000007c4055da5ad452d6901c94d65cb7dbc5ac7bbbaa"],"transactionHash":"0x8d121def15c5e15d47fde151e0feb8a418cbe73dd63a164289f13e1fac156c2d","transactionIndex":"0x0"}]`

	if err := json.Unmarshal([]byte(testlog), &logs); err != nil {
		t.Fatalf("failed to unmarshal test log: %s", err)
	}

	for _, log := range logs {
		res, err := handler(contractAddr, &log)
		if err != nil {
			continue
		}

		t.Logf("handled event: %v", res)
	}
}

func handler(contractAddr ethcommon.Address, log *ethtypes.Log) (interface{}, error) {
	delegationManager, err := pelldelegationmanager.NewPellDelegationManagerFilterer(contractAddr, bind.ContractFilterer(nil))
	if err != nil {
		return nil, err
	}

	// operator registered event
	if operatorRegisteredEvent, err := delegationManager.ParseOperatorRegistered(*log); err == nil {
		if !strings.EqualFold(operatorRegisteredEvent.Raw.Address.Hex(), contractAddr.Hex()) {
			return nil, fmt.Errorf("ParseEvent: event address %s does not match delegation manager %s",
				operatorRegisteredEvent.Raw.Address.Hex(), contractAddr.Hex())
		}

		return operatorRegisteredEvent, nil
	}

	// operator details modified event
	if operatorDetailsModifiedEvent, err := delegationManager.ParseOperatorDetailsModified(*log); err == nil {
		if !strings.EqualFold(operatorDetailsModifiedEvent.Raw.Address.Hex(), contractAddr.Hex()) {
			return nil, fmt.Errorf("ParseEvent: event address %s does not match delegation manager %s",
				operatorDetailsModifiedEvent.Raw.Address.Hex(), contractAddr.Hex())
		}

		return operatorDetailsModifiedEvent, nil
	}

	// operator metadata URI updated event
	if operatorMetadataURIUpdatedEvent, err := delegationManager.ParseOperatorMetadataURIUpdated(*log); err == nil {
		if !strings.EqualFold(operatorMetadataURIUpdatedEvent.Raw.Address.Hex(), contractAddr.Hex()) {
			return nil, fmt.Errorf("ParseEvent: event address %s does not match delegation manager %s",
				operatorMetadataURIUpdatedEvent.Raw.Address.Hex(), contractAddr.Hex())
		}

		return operatorMetadataURIUpdatedEvent, nil
	}

	return nil, nil
}
