package jsonrpc

import (
	"encoding/json"
	"fmt"
	"github.com/parnurzeal/gorequest"
	"time"
)

type Client struct {
	c *gorequest.SuperAgent
}

// NewEthClient 获取eth客户端
func NewEthClient(rawURI string) *Client {
	c := gorequest.New().
		Timeout(time.Minute * 5).
		Post(rawURI)
	return &Client{c}
}

// StRpcRespError rpc 错误
type StRpcRespError struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

func (e *StRpcRespError) Error() string {
	return fmt.Sprintf("%d %s", e.Code, e.Message)
}

// StEthTransaction 交易
type StEthTransaction struct {
	From             string      `json:"from"`
	Gas              string      `json:"gas"`
	GasPrice         string      `json:"gasPrice"`
	Hash             string      `json:"hash"`
	Input            string      `json:"input"`
	Nonce            string      `json:"nonce"`
	R                string      `json:"r"`
	S                string      `json:"s"`
	To               string      `json:"to"`
	TransactionIndex interface{} `json:"transactionIndex"`
	Type             string      `json:"type"`
	V                string      `json:"v"`
	Value            string      `json:"value"`
}

// StRpcReq rpc请求
type StRpcReq struct {
	Jsonrpc string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

// StRpcResp rpc返回
type StRpcResp struct {
	ID    string          `json:"id"`
	Error *StRpcRespError `json:"error"`
}

// doReq 发送请求
func (c *Client) doReq(method string, arqs []interface{}, resp interface{}) error {
	_, body, errs := c.c.
		Send(StRpcReq{
			Jsonrpc: "1.0",
			ID:      "1",
			Method:  method,
			Params:  arqs,
		}).EndBytes()
	if errs != nil {
		return errs[0]
	}
	err := json.Unmarshal(body, resp)
	if err != nil {
		return err
	}
	return nil
}

// EthRpcNetVersion 获取block信息
// "1": Ethereum Mainnet
// "2": Morden Testnet (deprecated)
// "3": Ropsten Testnet
// "4": Rinkeby Testnet
// "42": Kovan Testnet
func (c *Client) EthRpcNetVersion() (int64, error) {
	resp := struct {
		StRpcResp
		Result int64 `json:"result,string"`
	}{}
	err := c.doReq(
		"net_version",
		nil,
		&resp,
	)
	if err != nil {
		return 0, err
	}
	if resp.Error != nil {
		return 0, resp.Error
	}
	return resp.Result, nil
}

func (c *Client) TxPoolContent() (map[string]map[string]map[string]*StEthTransaction, error) {
	resp := struct {
		StRpcResp
		Result map[string]map[string]map[string]*StEthTransaction `json:"result"`
	}{}
	err := c.doReq(
		"txpool_content",
		[]interface{}{},
		&resp,
	)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, resp.Error
	}
	return resp.Result, nil
}

// EthRpcSendRawTransaction 发送交易
func (c *Client) EthRpcSendRawTransaction(rawTx string) (string, error) {
	resp := struct {
		StRpcResp
		Result string `json:"result"`
	}{}
	err := c.doReq(
		"eth_sendRawTransaction",
		[]interface{}{
			rawTx,
		},
		&resp,
	)
	if err != nil {
		return "", err
	}
	if resp.Error != nil {
		return "", resp.Error
	}
	return resp.Result, nil
}
