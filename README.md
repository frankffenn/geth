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
   --fromKey value   specify the private key of the send wallet
   --fromCode value  specify the password of the send wallet
   --toKey value     the public key of the receive wallet
   --amount value    the amount of token (0.00001 eth) (default: 1000)
   --help, -h        show help (default: false)
```

Example:

transfer eth
```
geth eth --fromKey=YourPrivateKey --fromCode=YourPassword --toKey=ToAddress --amount=1000
```
or bzz

```
geth bzz --fromKey=YourPrivateKey --fromCode=YourPassword --toKey=ToAddress --amount=1000
```

more token will be support

# license
MIT