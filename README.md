# geth manual

use `geth-cli` tool to transfer `ETH` and `gBZZ` token 

Usage:
```
~/go/src/geth-cli » geth-cli help
NAME:
   geth-cli - Common Ethereum tools

USAGE:
   geth-cli [global options] command [command options] [arguments...]

COMMANDS:
   txpool     
   gas-price  return the current gas price (Gwei)
   eth        
   bzz        
   help, h    Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help (default: false)
```

```
~/gopath/src/geth » geth-cli eth help   
NAME:
   geth-cli eth - A new cli application

USAGE:
   geth-cli eth command [command options] [arguments...]

COMMANDS:
   bls      
   send     
   help, h  Shows a list of commands or help for one command

OPTIONS:
   --help, -h  show help (default: false)
```

Example:

query current eth gas price
```
./geth-cli gas-price

 currnet gasPrice: 115.2 Gwei
```

transfer eth
```
./geth-cli eth send --fromKey=yourPrivateKey --toKey=toAddress --amount=100000 --gasLimit=1000000 --nGasPrice=2
```
or bzz

```
./geth-cli bzz send --fromKey=yourPrivateKey --toKey=toAddress --amount=100000 --gasLimit=3000000 --nGasPrice=2
```

replace gas price from txpool
```
./geth-cli txpool replace --from=youtWalletAddress --fromKey=yourPrivateKey --nGasPrice=2 --gasLimit=1
```

more token will be support

# license
MIT
