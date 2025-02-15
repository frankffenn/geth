package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"geth-cli/jsonrpc"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
)

var defaultEndPoint = "http://120.79.149.59:8545"

var client *jsonrpc.Client

func main() {
	endpoint := os.Getenv("ENDPOINT")
	if endpoint != "" {
		defaultEndPoint = endpoint
	}

	client = jsonrpc.NewEthClient(defaultEndPoint)

	local := []*cli.Command{
		txPoolCmd,
		gasPriceCmd,
		ETHCmd,
		BZZCmd,
	}

	app := &cli.App{
		Name:     "geth-cli",
		Usage:    "Common Ethereum tools",
		Commands: local,
	}

	if err := app.Run(os.Args); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

var gasPriceCmd = &cli.Command{
	Name:  "gas-price",
	Usage: "return the current gas price (Gwei)",
	Action: func(c *cli.Context) error {
		client, err := ethclient.Dial(defaultEndPoint)
		if err != nil {
			return err
		}

		gasPrice, err := client.SuggestGasPrice(context.Background())
		if err != nil {
			return err
		}

		fmt.Printf("currnet gasPrice: %v Gwei\n", float64(gasPrice.Int64())/1000000000)
		return nil
	},
}

var ethBalancesCmd = &cli.Command{
	Name: "bls",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "address",
			Value:    "",
			Required: true,
			Usage:    "the ethereum address",
		},
	},
	Action: func(c *cli.Context) error {
		client, err := ethclient.Dial(defaultEndPoint)
		if err != nil {
			return err
		}
		account := common.HexToAddress(c.String("address"))
		balance, err := client.BalanceAt(context.Background(), account, nil)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(balance)
		return nil
	},
}

var sendEthCmd = &cli.Command{
	Name: "send",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "fromKey",
			Value:    "",
			Required: true,
			Usage:    "specify the private key of the send wallet",
		},
		&cli.StringFlag{
			Name:     "toKey",
			Value:    "",
			Required: true,
			Usage:    "the public key of the receive wallet",
		},
		&cli.Int64Flag{
			Name:     "amount",
			Value:    1000,
			Required: true,
			Usage:    "the amount of token (0.00001 eth)",
		},
		&cli.Uint64Flag{
			Name:  "gasLimit",
			Value: 21000, // in units
			Usage: "the amount of gas limit (wei)",
		},
		&cli.Uint64Flag{
			Name:  "nGasPrice",
			Value: 2, // in units
			Usage: "n times of the current gas price",
		},
	},
	Action: func(c *cli.Context) error {
		return PayEth(defaultEndPoint, c.String("fromKey"), c.String("toKey"), c.Int64("amount"), c.Uint64("gasLimit"), c.Uint64("nGasPrice"))
	},
}

var ETHCmd = &cli.Command{
	Name: "eth",
	Subcommands:[]*cli.Command{
		ethBalancesCmd,
		sendEthCmd,
	},
}

// PayEth 传入出账的私钥（geth导入账号那部分），传入要进账的公钥，金额单位是wei * 10的9次方。
func PayEth(endpoint, fromKey, toKey string, amount int64, gasLimit, nGasPrice uint64) error {
	if toKey == "" {
		return xerrors.New("receiver must not be empty")
	}

	client, err := ethclient.Dial(endpoint)
	if err != nil {
		return err
	}

	// 出账的私钥。
	privateKey, err := crypto.HexToECDSA(fromKey)
	if err != nil {
		return err
	}

	// 出账的公钥
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Println("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
		return err
	}

	// 进账的地址。
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return err
	}

	// 支付的金额。
	value := big.NewInt(amount * 100000000000000) // in wei (0.000000001 eth)
	price, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}

	gasPrice := big.NewInt(0).Mul(price, big.NewInt(int64(nGasPrice)))

	log.Println("gas Price:", gasPrice)
	log.Println("gas Limit:", gasLimit)

	toAddress := common.HexToAddress(toKey)
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return err
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return err
	}

	log.Printf("eth tx sent: %s", signedTx.Hash().Hex())

	return nil
}
