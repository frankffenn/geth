package main

import (
	"context"
	"encoding/json"
	"fmt"
	"geth-cli/jsonrpc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/urfave/cli/v2"
	"log"
	"math/big"
	"strings"
	"time"
)

var txPoolCmd = &cli.Command{
	Name: "txpool",
	Subcommands: []*cli.Command{
		pendingCmd,
		replaceCmd,
	},
}

var pendingCmd = &cli.Command{
	Name: "pending",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "from",
			Value: "",
			Usage: "filters the specified wallet address",
		},
	},
	Action: func(c *cli.Context) error {
		transactions, err := client.TxPoolContent()
		if err != nil {
			log.Fatal("txpool:", err)
		}

		filter := c.String("from")
		pending := transactions["pending"]

		var out []*jsonrpc.StEthTransaction
		for _, pv := range pending {
			for _, v := range pv {
				if filter != "" && filter != v.From {
					continue
				}
				out = append(out, v)
			}
		}

		if out == nil {
			return nil
		}

		bytes, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(bytes))
		return nil
	},
}

var replaceCmd = &cli.Command{
	Name: "replace",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "from",
			Value:    "",
			Required: true,
			Usage:    "filters the specified wallet address",
		},
		&cli.StringFlag{
			Name:     "fromKey",
			Value:    "",
			Required: true,
			Usage:    "the private key for the replace messages",
		},
		&cli.Uint64Flag{
			Name:  "nGasPrice",
			Value: 2, // in units
			Usage: "the amount of gas price (wei)",
		},
		&cli.Uint64Flag{
			Name:  "gasLimit",
			Value: 0, // in units
			Usage: "the amount of gas limit (wei)",
		},
	},
	Action: func(c *cli.Context) error {
		allTransactions, err := client.TxPoolContent()
		if err != nil {
			log.Fatal("txpool:", err)
		}

		filter := c.String("from")
		pending := allTransactions["pending"]

		var transactions []*jsonrpc.StEthTransaction
		for _, pv := range pending {
			for _, v := range pv {
				if filter != "" && filter != v.From {
					continue
				}
				transactions = append(transactions, v)
			}
		}

		if transactions == nil {
			return nil
		}

		netVer, err := client.EthRpcNetVersion()
		if err != nil {
			log.Panicf("error net ver: %s", err.Error())
		}

		log.Printf("netVer: %d\n", netVer)

		fromKey := c.String("fromKey")
		if strings.HasPrefix(fromKey, "0x") {
			fromKey = fromKey[2:]
		}

		privateKey, err := crypto.HexToECDSA(fromKey)
		if err != nil {
			log.Fatalf("err: [%T] %s", err, err.Error())
		}

		for _, rpcTx := range transactions {
			etchClient, err := ethclient.Dial(defaultEndPoint)
			if err != nil {
				return err
			}

			gasPriceTimes := c.Uint64("nGasPrice")
			gasPrice, err := etchClient.SuggestGasPrice(context.Background())
			if err != nil {
				return err
			}

			gasPrice = gasPrice.Mul(gasPrice, big.NewInt(int64(gasPriceTimes)))

			oldGasPrice, err := hexutil.DecodeBig(rpcTx.GasPrice)
			if err != nil {
				log.Fatalf("error gas price: %s", rpcTx.GasPrice)
			}

			if gasPrice.Cmp(oldGasPrice) < 0 {
				log.Println("Message does not need replace")
				continue
			}

			var inputBs []byte
			if len(rpcTx.Input) > 0 {
				inputBs, err = hexutil.Decode(rpcTx.Input)
				if err != nil {
					log.Fatalf("input decode err: %s", err.Error())
				}
			}

			gasLimit, err := hexutil.DecodeUint64(rpcTx.Gas)
			if err != nil {
				log.Fatalf("error gas limit: %s", rpcTx.Gas)
			}

			if c.Uint64("gasLimit") > gasLimit {
				gasLimit = c.Uint64("gasLimit")
			}

			nonce, err := hexutil.DecodeUint64(rpcTx.Nonce)
			if err != nil {
				log.Fatalf("error nonce: %s", rpcTx.Nonce)
			}
			value, err := hexutil.DecodeBig(rpcTx.Value)
			if err != nil {
				log.Fatalf("error value: %s", rpcTx.Value)
			}

			ethTx := types.NewTransaction(
				nonce,
				common.HexToAddress(rpcTx.To),
				value,
				gasLimit,
				gasPrice,
				inputBs,
			)

			// 签名
			signedTx, err := types.SignTx(ethTx, types.NewEIP155Signer(big.NewInt(netVer)), privateKey)
			if err != nil {
				log.Fatalf("sign tx err: %s", err.Error())
			}
			ts, err := rlp.EncodeToBytes(signedTx)
			if err != nil {
				log.Fatalf("err encode tx: %s", err.Error())
			}

			rawTxHex := hexutil.Encode(ts)
			txHash := strings.ToLower(signedTx.Hash().Hex())
			log.Printf("tx: %s\n hex:\n%s\n", txHash, rawTxHex)
			sendTxID, err := client.EthRpcSendRawTransaction(rawTxHex)
			if err != nil {
				log.Printf("send tx err: %s\n", err.Error())
				continue
			}

			time.Sleep(1 *time.Second)
			log.Printf("send result: %s\n", sendTxID)
		}

		return nil
	},
}
