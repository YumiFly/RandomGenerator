package events

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"

	"RandomGenerator/VRFCoordinatorV2"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	infuraURL = "ws://127.0.0.1:8545/"
)

// var contractABI = `[{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"requestId","type":"uint256"},{"indexed":true,"internalType":"address","name":"requester","type":"address"},{"indexed":false,"internalType":"uint256","name":"subId","type":"uint256"},{"indexed":false,"internalType":"bytes32","name":"keyHash","type":"bytes32"},{"indexed":false,"internalType":"uint32","name":"callbackGasLimit","type":"uint32"},{"indexed":false,"internalType":"uint16","name":"requestConfirmations","type":"uint16"},{"indexed":false,"internalType":"uint32","name":"numWords","type":"uint32"},{"indexed":false,"internalType":"bytes","name":"extraArgs","type":"bytes"}],"name":"RandomWordsRequested","type":"event"}]`
// var contractABI = `[{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"requestId","type":"uint256"},{"indexed":true,"internalType":"address","name":"requester","type":"address"},{"indexed":false,"internalType":"uint256","name":"subId","type":"uint256"},{"indexed":false,"internalType":"bytes32","name":"keyHash","type":"bytes32"},{"indexed":false,"internalType":"uint32","name":"callbackGasLimit","type":"uint32"},{"indexed":false,"internalType":"uint16","name":"requestConfirmations","type":"uint16"},{"indexed":false,"internalType":"uint32","name":"numWords","type":"uint32"},{"indexed":false,"internalType":"bytes","name":"extraArgs","type":"bytes"}],"name":"RandomWordsRequested","type":"event"},{"inputs":[{"internalType":"uint256","name":"requestId","type":"uint256"},{"internalType":"uint256[]","name":"randomWords","type":"uint256[]"}],"name":"CallFullfillRandomWords","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"components":[{"internalType":"bytes32","name":"keyHash","type":"bytes32"},{"internalType":"uint256","name":"subId","type":"uint256"},{"internalType":"uint16","name":"requestConfirmations","type":"uint16"},{"internalType":"uint32","name":"callbackGasLimit","type":"uint32"},{"internalType":"uint32","name":"numWords","type":"uint32"},{"internalType":"bytes","name":"extraArgs","type":"bytes"}],"internalType":"structVRFV2PlusClient.RandomWordsRequest","name":"req","type":"tuple"}],"name":"requestRandomWords","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"","type":"uint256"}],"name":"s_requesters","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"}]`
var contractABI = `[
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "requestId",
				"type": "uint256"
			},
			{
				"indexed": false,
				"internalType": "address",
				"name": "requester",
				"type": "address"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "subId",
				"type": "uint256"
			},
			{
				"indexed": false,
				"internalType": "bytes32",
				"name": "keyHash",
				"type": "bytes32"
			},
			{
				"indexed": false,
				"internalType": "uint32",
				"name": "callbackGasLimit",
				"type": "uint32"
			},
			{
				"indexed": false,
				"internalType": "uint16",
				"name": "requestConfirmations",
				"type": "uint16"
			},
			{
				"indexed": false,
				"internalType": "uint32",
				"name": "numWords",
				"type": "uint32"
			},
			{
				"indexed": false,
				"internalType": "bytes",
				"name": "extraArgs",
				"type": "bytes"
			}
		],
		"name": "RandomWordsRequested",
		"type": "event"
	},
	{
		"inputs": [
			{
				"internalType": "uint256",
				"name": "requestId",
				"type": "uint256"
			},
			{
				"internalType": "uint256[]",
				"name": "randomWords",
				"type": "uint256[]"
			}
		],
		"name": "CallFullfillRandomWords",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"components": [
					{
						"internalType": "bytes32",
						"name": "keyHash",
						"type": "bytes32"
					},
					{
						"internalType": "uint256",
						"name": "subId",
						"type": "uint256"
					},
					{
						"internalType": "uint16",
						"name": "requestConfirmations",
						"type": "uint16"
					},
					{
						"internalType": "uint32",
						"name": "callbackGasLimit",
						"type": "uint32"
					},
					{
						"internalType": "uint32",
						"name": "numWords",
						"type": "uint32"
					},
					{
						"internalType": "bytes",
						"name": "extraArgs",
						"type": "bytes"
					}
				],
				"internalType": "struct VRFV2PlusClient.RandomWordsRequest",
				"name": "req",
				"type": "tuple"
			}
		],
		"name": "requestRandomWords",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"name": "s_requesters",
		"outputs": [
			{
				"internalType": "address",
				"name": "",
				"type": "address"
			}
		],
		"stateMutability": "view",
		"type": "function"
	}
]`

type EventListener struct {
	client          *ethclient.Client
	contractAddress common.Address
}

type EventData struct {
	RequestId            *big.Int
	Requester            common.Address
	SubId                *big.Int
	KeyHash              [32]byte
	CallbackGasLimit     uint32
	RequestConfirmations uint16
	NumWords             uint32
	ExtraArgs            []byte
}

func NewEventListener(contractAddress string) (*EventListener, error) {
	client, err := ethclient.Dial(infuraURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the Ethereum client: %v", err)
	}

	return &EventListener{
		client:          client,
		contractAddress: common.HexToAddress(contractAddress),
	}, nil
}

func (el *EventListener) Listen() {
	query := ethereum.FilterQuery{
		Addresses: []common.Address{el.contractAddress},
	}

	logs := make(chan types.Log)
	sub, err := el.client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Fatalf("Failed to subscribe to logs: %v", err)
	}

	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	for {
		select {
		case err := <-sub.Err():
			log.Fatalf("Error: %v", err)
		case vLog := <-logs:
			event := EventData{}
			err := parsedABI.UnpackIntoInterface(&event, "RandomWordsRequested", vLog.Data)
			if err == nil {
				// 调用 Process 方法
				el.Process(&event)
			}
		}
	}
}

func (el *EventListener) Process(event *EventData) {
	requestId := event.RequestId

	var rndNumbers []*big.Int
	for i := 0; i < int(event.NumWords); i++ {
		// 生成 256 位的随机数
		maxInt := new(big.Int).Lsh(big.NewInt(1), 256)
		randomNumber, err := rand.Int(rand.Reader, maxInt)
		if err != nil {
			log.Fatalf("Failed to generate random number: %v", err)
		}
		rndNumbers = append(rndNumbers, randomNumber)
	}

	instance, err := VRFCoordinatorV2.NewVRFCoordinatorV2(el.contractAddress, el.client)
	if err != nil {
		log.Fatalf("Failed to instantiate a VRFCoordinatorV2 contract: %v", err)
	}

	// 使用私钥生成 ECDSA 密钥对
	privKey, err := crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	if err != nil {
		log.Fatalf("failed to get private key")
	}
	// 获取当前账户的地址
	pubAddr := crypto.PubkeyToAddress(privKey.PublicKey)
	// 获取当前账户的nonce值
	nonce, err := el.client.PendingNonceAt(context.Background(), pubAddr)
	if err != nil {
		log.Fatalf("failed to get nonce")
	}
	// 获取当前推荐的gasPrice
	gasPrice, err := el.client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf("failed to get gas price")
	}
	gasLimit := uint64(30000000)

	// 调用 VRFCoordinatorV2 包中的 CallFullfillRandomWords 方法
	Opts := &bind.TransactOpts{
		From:  pubAddr,
		Nonce: big.NewInt(int64(nonce)),
		Signer: func(addr common.Address, tx *types.Transaction) (*types.Transaction, error) {
			chainId, err := el.client.NetworkID(context.Background())
			if err != nil {
				return nil, errors.New("failed to get network ID")
			}
			return types.SignTx(tx, types.NewEIP155Signer(chainId), privKey)
		},
		Value:    nil,
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		Context:  context.Background(),
	}
	_, err = instance.CallFullfillRandomWords(Opts, requestId, rndNumbers)
	if err != nil {
		log.Fatalf("Failed to call CallFullfillRandomWords: %v", err)
	}
	for i, it := range rndNumbers {
		fmt.Printf("Successfully called CallFullfillRandomWords with RequestId: %s and RandomNumber[%d]: %s\n", requestId.String(), i, it.String())
	}
}
