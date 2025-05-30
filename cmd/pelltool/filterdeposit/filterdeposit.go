package filterdeposit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/aegis/cmd/pelltool/config"
	"github.com/0xPellNetwork/aegis/x/relayer/types"
)

const (
	BTCChainIDFlag = "btc-chain-id"
)

func NewFilterDepositCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "filterdeposit",
		Short: "filter missing inbound deposits",
	}

	cmd.AddCommand(NewEvmCmd())

	// Required for TSS address query
	cmd.PersistentFlags().String(BTCChainIDFlag, "8332", "chain id used on pellchain to identify bitcoin - default: 8332")

	return cmd
}

// Deposit is a data structure for keeping track of inbound transactions
type Deposit struct {
	TxID   string
	Amount uint64
}

// CheckForXmsg is querying pell core for a xmsg associated with a confirmed transaction hash. If the xmsg is not found,
// then the transaction hash is added to the list of missed inbound transactions.
func CheckForXmsg(list []Deposit, cfg *config.Config) ([]Deposit, error) {
	var missedList []Deposit

	fmt.Println("Going through list, num of transactions: ", len(list))
	for _, entry := range list {
		pellURL, err := url.JoinPath(cfg.PellURL, "pell-chain", "crosschain", "in_tx_hash_to_xmsg_data", entry.TxID)
		if err != nil {
			return missedList, err
		}

		request, err := http.NewRequest(http.MethodGet, pellURL, nil)
		if err != nil {
			return missedList, err
		}
		request.Header.Add("Accept", "application/json")
		client := &http.Client{}

		response, getErr := client.Do(request)
		if getErr != nil {
			return missedList, getErr
		}

		data, readErr := ioutil.ReadAll(response.Body)
		if readErr != nil {
			return missedList, readErr
		}
		closeErr := response.Body.Close()
		if closeErr != nil {
			return missedList, closeErr
		}

		var xmsg map[string]interface{}
		err = json.Unmarshal(data, &xmsg)
		if err != nil {
			return missedList, err
		}

		// successful query of the given xmsg will not contain a "message" field with value "not found", if it was not
		// found then it is added to the missing list.
		if _, ok := xmsg["message"]; ok {
			if strings.Compare(xmsg["message"].(string), "not found") == 0 {
				missedList = append(missedList, entry)
			}
		}
	}

	fmt.Printf("Found %d missed transactions.\n", len(missedList))
	for _, entry := range missedList {
		fmt.Printf("%s, amount: %d\n", entry.TxID, entry.Amount)
	}
	return missedList, nil
}

func GetTssAddress(cfg *config.Config, btcChainID string) (*types.QueryGetTssAddressResponse, error) {
	res := &types.QueryGetTssAddressResponse{}
	requestURL, err := url.JoinPath(cfg.PellURL, "pell-chain", "relayer", "get_tss_address", btcChainID)
	if err != nil {
		return res, err
	}
	request, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return res, err
	}
	request.Header.Add("Accept", "application/json")
	pellcoreHTTPClient := &http.Client{}
	response, getErr := pellcoreHTTPClient.Do(request)
	if getErr != nil {
		return res, err
	}
	data, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		return res, err
	}
	closeErr := response.Body.Close()
	if closeErr != nil {
		return res, closeErr
	}
	err = json.Unmarshal(data, res)
	return res, err
}
