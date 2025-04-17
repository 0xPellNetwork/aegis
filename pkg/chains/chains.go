package chains

import "fmt"

func (c Chain) ChainName() string {
	return fmt.Sprintf("%s-%s-%d", c.Network.String(), c.NetworkType.String(), c.Id)
}

// ChainsList returns a list of default chains
func ChainsList() []Chain {
	return []Chain{
		{
			Id:          86,
			Network:     NetWork_PELL,
			NetworkType: NetWorkType_MAINNET,
			VmType:      VMType_EVM,
		},
		{
			Id:          87,
			Network:     NetWork_PELL,
			NetworkType: NetWorkType_TESTNET,
			VmType:      VMType_EVM,
		},
		{
			Id:          860,
			Network:     NetWork_PELL,
			NetworkType: NetWorkType_PRIVNET,
			VmType:      VMType_EVM,
		},
		{
			Id:          186,
			Network:     NetWork_PELL,
			NetworkType: NetWorkType_PRIVNET,
			VmType:      VMType_EVM,
		},
		{
			Id:          1,
			Network:     NetWork_ETH,
			NetworkType: NetWorkType_MAINNET,
			VmType:      VMType_EVM,
		},
		{
			Id:          56,
			Network:     NetWork_BSC,
			NetworkType: NetWorkType_MAINNET,
			VmType:      VMType_EVM,
		},
		{
			Id:          137,
			Network:     NetWork_POLYGON,
			NetworkType: NetWorkType_MAINNET,
			VmType:      VMType_EVM,
		},
		{
			Id:          8453,
			Network:     NetWork_BASE,
			NetworkType: NetWorkType_MAINNET,
			VmType:      VMType_EVM,
		},
		{
			Id:          11155111,
			Network:     NetWork_SEPOLIA,
			NetworkType: NetWorkType_TESTNET,
			VmType:      VMType_EVM,
		},
		{
			Id:          5,
			Network:     NetWork_GOERLI,
			NetworkType: NetWorkType_TESTNET,
			VmType:      VMType_EVM,
		},
		{
			Id:          97,
			Network:     NetWork_BSC,
			NetworkType: NetWorkType_TESTNET,
			VmType:      VMType_EVM,
		},
		{
			Id:          18332,
			Network:     NetWork_BTC,
			NetworkType: NetWorkType_TESTNET,
			VmType:      VMType_NO_VM,
		},
		{
			Id:          80001,
			Network:     NetWork_MUMBAI,
			NetworkType: NetWorkType_TESTNET,
			VmType:      VMType_EVM,
		},
		{
			Id:          80002,
			Network:     NetWork_AMOY,
			NetworkType: NetWorkType_TESTNET,
			VmType:      VMType_EVM,
		},
		{
			Id:          84532,
			Network:     NetWork_SEPOLIA,
			NetworkType: NetWorkType_TESTNET,
			VmType:      VMType_EVM,
		},
		{
			Id:          18444,
			Network:     NetWork_BTC,
			NetworkType: NetWorkType_PRIVNET,
			VmType:      VMType_NO_VM,
		},
		{
			Id:          1337,
			Network:     NetWork_ETH,
			NetworkType: NetWorkType_PRIVNET,
			VmType:      VMType_EVM,
		},
		{
			Id:          5003,
			Network:     NetWork_MANTLE,
			NetworkType: NetWorkType_TESTNET,
			VmType:      VMType_EVM,
		},
		{
			Id:          5000,
			Network:     NetWork_MANTLE,
			NetworkType: NetWorkType_MAINNET,
			VmType:      VMType_EVM,
		},
		{
			Id:          1115,
			Network:     NetWork_CORE,
			NetworkType: NetWorkType_TESTNET,
			VmType:      VMType_EVM,
		},
		{
			Id:          1116,
			Network:     NetWork_CORE,
			NetworkType: NetWorkType_MAINNET,
			VmType:      VMType_EVM,
		},
		{
			Id:          200810,
			Network:     NetWork_BITLAYER,
			NetworkType: NetWorkType_TESTNET,
			VmType:      VMType_EVM,
		}, {
			Id:          200901,
			Network:     NetWork_BITLAYER,
			NetworkType: NetWorkType_MAINNET,
			VmType:      VMType_EVM,
		},
		{
			Id:          42161,
			Network:     NetWork_ARBITRUM,
			NetworkType: NetWorkType_MAINNET,
			VmType:      VMType_EVM,
		},
		{
			Id:          421614,
			Network:     NetWork_ARBITRUM,
			NetworkType: NetWorkType_TESTNET,
			VmType:      VMType_EVM,
		},
		{
			Id:          42170,
			Network:     NetWork_ARBITRUM,
			NetworkType: NetWorkType_TESTNET,
			VmType:      VMType_EVM,
		},
	}
}

func FindChains(filter func(Chain) bool) []Chain {
	res := make([]Chain, 0)
	for _, v := range ChainsList() {
		if filter(v) {
			res = append(res, v)
		}
	}

	return res
}

func FindChain(filter func(Chain) bool) (Chain, bool) {
	for _, v := range ChainsList() {
		if filter(v) {
			return v, true
		}
	}

	return Chain{}, false
}

func GetChainByChainId(id int64) (Chain, bool) {
	return FindChain(func(c Chain) bool { return c.Id == id })
}

func PrivnetChainList() []Chain {
	return FindChains(func(c Chain) bool { return c.NetworkType == NetWorkType_PRIVNET })
}

func ExternalChainList() []Chain {
	return FindChains(func(c Chain) bool { return c.Network != NetWork_PELL })
}

// PellChainFromChainID returns a PellChain chainobject  from a Cosmos chain ID
func PellChainFromChainID(chainID string) (Chain, error) {
	ethChainID, err := CosmosToEthChainID(chainID)
	if err != nil {
		return Chain{}, err
	}

	chain, exist := FindChain(func(c Chain) bool {
		return c.Id == ethChainID
	})

	if !exist {
		return Chain{}, fmt.Errorf("chain %d not found", ethChainID)
	}

	return chain, nil
}

// chain simple define
func GoerliChain() Chain {
	return Chain{
		Id:          5,
		Network:     NetWork_GOERLI,
		NetworkType: NetWorkType_TESTNET,
		VmType:      VMType_EVM,
	}
}

func BscMainnetChain() Chain {
	return Chain{
		Id:          56,
		Network:     NetWork_BSC,
		NetworkType: NetWorkType_MAINNET,
		VmType:      VMType_EVM,
	}
}

func EthChain() Chain {
	return Chain{
		Id:          1,
		Network:     NetWork_ETH,
		NetworkType: NetWorkType_MAINNET,
		VmType:      VMType_EVM,
	}
}

func GoerliLocalnetChain() Chain {
	return Chain{
		Id:          1337,
		Network:     NetWork_ETH,
		NetworkType: NetWorkType_PRIVNET,
		VmType:      VMType_EVM,
	}
}

func PellChainMainnet() Chain {
	return Chain{
		Id:          86,
		Network:     NetWork_PELL,
		NetworkType: NetWorkType_MAINNET,
		VmType:      VMType_EVM,
	}
}

func SepoliaChain() Chain {
	return Chain{
		Id:          11155111,
		Network:     NetWork_SEPOLIA,
		NetworkType: NetWorkType_TESTNET,
		VmType:      VMType_EVM,
	}
}

func PellPrivnetChain() Chain {
	return Chain{
		Id:          186,
		Network:     NetWork_PELL,
		NetworkType: NetWorkType_PRIVNET,
		VmType:      VMType_EVM,
	}
}

func BscTestnetChain() Chain {
	return Chain{
		Id:          97,
		Network:     NetWork_BSC,
		NetworkType: NetWorkType_TESTNET,
		VmType:      VMType_EVM,
	}
}

func MumbaiChain() Chain {
	return Chain{
		Id:          80001,
		Network:     NetWork_MUMBAI,
		NetworkType: NetWorkType_TESTNET,
		VmType:      VMType_EVM,
	}
}

func PellTestnetChain() Chain {
	return Chain{
		Id:          87,
		Network:     NetWork_PELL,
		NetworkType: NetWorkType_TESTNET,
		VmType:      VMType_EVM,
	}
}

func CoreTestNetChain() Chain {
	return Chain{
		Id:          87,
		Network:     NetWork_PELL,
		NetworkType: NetWorkType_TESTNET,
		VmType:      VMType_EVM,
	}
}
