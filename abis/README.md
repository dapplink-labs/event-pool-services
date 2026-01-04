# Smart Contract ABIs

This directory contains your smart contract ABI (Application Binary Interface) files.

## How to Add Contract ABIs

1. Place your contract ABI JSON files in this directory
2. Organize by contract name (e.g., `MyContract.sol/MyContract.json`)
3. Use these ABI files to generate Go bindings

## Generate Go Bindings

Use the `abigen` tool from go-ethereum to generate Go bindings:

### Install abigen

```bash
go install github.com/ethereum/go-ethereum/cmd/abigen@latest
```

### Generate Bindings

```bash
# Example: Generate binding for a contract
abigen --abi abis/MyContract.sol/MyContract.json \
       --pkg bindings \
       --type MyContract \
       --out bindings/my_contract.go
```

### Directory Structure Example

```
abis/
├── MyContract.sol/
│   └── MyContract.json
├── AnotherContract.sol/
│   └── AnotherContract.json
└── README.md
```

Then generate bindings:

```bash
abigen --abi abis/MyContract.sol/MyContract.json --pkg bindings --type MyContract --out bindings/my_contract.go
abigen --abi abis/AnotherContract.sol/AnotherContract.json --pkg bindings --type AnotherContract --out bindings/another_contract.go
```

## Using Contract Bindings

After generating bindings, you can use them in your code:

```go
import (
    "github.com/ethereum/go-ethereum/ethclient"
    "your-project/bindings"
)

// Connect to Ethereum node
client, err := ethclient.Dial("https://your-rpc-url")
if err != nil {
    log.Fatal(err)
}

// Create contract instance
contractAddress := common.HexToAddress("0x...")
contract, err := bindings.NewMyContract(contractAddress, client)
if err != nil {
    log.Fatal(err)
}

// Call contract methods
result, err := contract.SomeMethod(nil)
```

## Resources

- [abigen Documentation](https://geth.ethereum.org/docs/tools/abigen)
- [go-ethereum Documentation](https://pkg.go.dev/github.com/ethereum/go-ethereum)
