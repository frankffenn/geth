# geth manual

use `geth` tool to transfer `ETH` and `gBZZ` token 

Usage:
```
~/go/src/geth » geth help
NAME:
   geth - Send ETH and ERC-20 token

USAGE:
   geth [global options] command [command options] [arguments...]

COMMANDS:
   eth
   bzz
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help (default: false)
```

```
~/gopath/src/geth » geth eth help   
NAME:
   geth eth -

USAGE:
   geth eth [command options] [arguments...]

OPTIONS:
   --fromKey value    specify the private key of the send wallet
   --toKey value      the public key of the receive wallet
   --amount value     the amount of token (0.00001 eth) (default: 1000)
   --gasLimit value   the amount of gas limit (wei) (default: 21000)
   --nGasPrice value  n times of the current gas price (default: 2)
   --help, -h         show help (default: false)
```

Example:

transfer eth
```
./geth eth --fromKey=yourPrivateKey  --toKey=toAddress --amount=5000 --nGasPrice=2
```
or bzz

```
./geth bzz --fromKey=yourPrivateKey --toKey=toAddress --amount=100000 --gasLimit=3000000 --nGasPrice=2
```

more token will be support

# license
MIT