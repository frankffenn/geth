package main

import (
	"context"
	"crypto/ecdsa"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/sha3"
)

const (
	defaultEndPoint = "https://goerli.infura.io/v3/f745456f30a8485babdd0bb3aa42725e"
)

func main() {
	local := []*cli.Command{
		sendETHCmd,
		sendBZZCmd,
	}

	app := &cli.App{
		Name:     "geth",
		Usage:    "Send ETH and ERC-20 token",
		Commands: local,
	}

	if err := app.Run(os.Args); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

var sendETHCmd = &cli.Command{
	Name: "eth",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "fromKey",
			Value:    "",
			Required: true,
			Usage:    "specify the private key of the send wallet",
		},
		&cli.StringFlag{
			Name:     "fromCode",
			Value:    "",
			Required: true,
			Usage:    "specify the password of the send wallet",
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
			Name:  "gasPrice",
			Value: 0, // in units
			Usage: "the amount of gas price (wei)",
		},
	},
	Action: func(c *cli.Context) error {
		endpoint := os.Getenv("ENDPOINT")
		if endpoint == "" {
			endpoint = defaultEndPoint
		}

		return PayEth(endpoint, c.String("fromKey"), c.String("fromCode"), c.String("toKey"), c.Int64("amount"), c.Uint64("gasLimit"), c.Uint64("gasPrice"))
	},
}

var sendBZZCmd = &cli.Command{
	Name: "bzz",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "fromKey",
			Value:    "",
			Required: true,
			Usage:    "specify the private key of the send wallet",
		},
		&cli.StringFlag{
			Name:     "fromCode",
			Value:    "",
			Required: true,
			Usage:    "specify the password of the send wallet",
		},
		&cli.StringFlag{
			Name:     "toKey",
			Value:    "",
			Required: true,
			Usage:    "the public key of the receive wallet",
		},
		&cli.Int64Flag{
			Name:     "amount",
			Value:    100000,
			Required: true,
			Usage:    "the amount of token (0.00001 gBzz)",
		},
		&cli.Uint64Flag{
			Name:  "gasLimit",
			Value: 0,
			Usage: "the amount of gas limit (wei)",
		},
		&cli.Uint64Flag{
			Name:  "gasPrice",
			Value: 0, // in units
			Usage: "the amount of gas price (wei)",
		},
	},
	Action: func(c *cli.Context) error {
		endpoint := os.Getenv("ENDPOINT")
		if endpoint == "" {
			endpoint = defaultEndPoint
		}

		return PayBzz(endpoint, c.String("fromKey"), c.String("fromCode"), c.String("toKey"), c.Int64("amount"), c.Uint64("gasLimit"), c.Uint64("gasPrice"))
	},
}

// PayEth 传入出账的私钥，密码（geth导入账号那部分），传入要进账的公钥，金额单位是wei * 10的9次方。
func PayEth(endpoint, fromKey, fromCode, toKey string, amount int64, gasLimit, gasPrice uint64) error {
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

	if big.NewInt(int64(gasPrice)).Cmp(price) > 0 {
		price = big.NewInt(int64(gasPrice))
	}

	log.Println("gas Price:", price)
	log.Println("gas Limit:", gasLimit)

	toAddress := common.HexToAddress(toKey)
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, price, nil)

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

// PayBzz 传入出账的私钥，密码（geth导入账号那部分），传入要进账的公钥，金额单位是wei * 10的9次方。
func PayBzz(endpoint, fromKey, fromCode, toKey string, amount int64, gasLimit, gasPrice uint64) error {
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

	// 代币传输不需要传输ETH，因此将交易“值”设置为“0”。
	value := big.NewInt(0)
	price, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}

	if big.NewInt(int64(gasPrice)).Cmp(price) > 0 {
		price = big.NewInt(int64(gasPrice))
	}

	toAddress := common.HexToAddress(toKey)
	var data []byte
	// 智能合约地址
	transferFnSignature := []byte("transfer(address,uint256)")
	tokenAddress := common.HexToAddress("0x2ac3c1d3e24b45c6c310534bc2dd84b5ed576335")

	// 生成函数签名
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4]

	// 将给我们发送代币的地址左填充到32字节
	paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)

	sentAmount := big.NewInt(amount * 1000000000000)
	paddedAmount := common.LeftPadBytes(sentAmount.Bytes(), 32)

	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)

	// 获取燃气上限制
	gLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
		To:   &toAddress,
		Data: data,
	})
	if err != nil {
		log.Fatal(err)
	}

	if gLimit > gasLimit {
		gasLimit = gLimit
	}

	log.Println("gas Price:", price)
	log.Println("gas Limit:", gasLimit)

	tx := types.NewTransaction(nonce, tokenAddress, value, gasLimit, price, data)

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

	log.Printf("bzz tx sent: %s", signedTx.Hash().Hex())

	return nil

}
