package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"geth-cli/erc20-token"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"math/big"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/sha3"
	"golang.org/x/xerrors"
	"log"
)

const bzzTokenAddress  = "0x2ac3c1d3e24b45c6c310534bc2dd84b5ed576335"

var BZZCmd = &cli.Command{
	Name: "bzz",
	Subcommands:[]*cli.Command{
		bzzBalancesCmd,
		sendBzzCmd,
	},
}

var bzzBalancesCmd = &cli.Command{
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
		tokenAddress := common.HexToAddress(bzzTokenAddress)
		instance, err := token.NewToken(tokenAddress, client)
		if err != nil {
			log.Fatal(err)
		}

		address := common.HexToAddress(c.String("address"))
		bal, err := instance.BalanceOf(&bind.CallOpts{}, address)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(bal)
		return nil
	},
}

var sendBzzCmd  = &cli.Command{
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
			Name:  "nGasPrice",
			Value: 2, // in units
			Usage: "n times of the current gas price",
		},
	},
	Action: func(c *cli.Context) error {
		return PayBzz(defaultEndPoint, c.String("fromKey"), c.String("toKey"), c.Int64("amount"), c.Uint64("gasLimit"), c.Uint64("nGasPrice"))
	},
}

// PayBzz 传入出账的私钥（geth导入账号那部分），传入要进账的公钥，金额单位是wei * 10的9次方。
func PayBzz(endpoint, fromKey, toKey string, amount int64, gasLimit, nGasPrice uint64) error {
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

	// 代币传输不需要传输ETH，因此将交易“值”设置为“0”。
	value := big.NewInt(0)
	price, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}

	gasPrice := big.NewInt(0).Mul(price, big.NewInt(int64(nGasPrice)))

	toAddress := common.HexToAddress(toKey)
	var data []byte
	// 智能合约地址
	transferFnSignature := []byte("transfer(address,uint256)")
	tokenAddress := common.HexToAddress(bzzTokenAddress)

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

	log.Println("gas Price:", gasPrice)
	log.Println("gas Limit:", gasLimit)

	tx := types.NewTransaction(nonce, tokenAddress, value, gasLimit, gasPrice, data)

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