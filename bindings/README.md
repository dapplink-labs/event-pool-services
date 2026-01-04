# Smart Contract Go Bindings

This directory contains Go bindings generated from smart contract ABIs.

## How to Generate Bindings

Use the `abigen` tool to generate Go bindings from your ABI files:

```bash
# Install abigen
go install github.com/ethereum/go-ethereum/cmd/abigen@latest

# Generate binding
abigen --abi abis/YourContract.sol/YourContract.json \
       --pkg bindings \
       --type YourContract \
       --out bindings/your_contract.go
```

## Example Usage

After generating bindings, import and use them in your code:

```go
package main

import (
    "log"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "your-project/bindings"
)

func main() {
    // Connect to Ethereum node
    client, err := ethclient.Dial("https://mainnet.infura.io/v3/YOUR-PROJECT-ID")
    if err != nil {
        log.Fatal(err)
    }

    // Contract address
    address := common.HexToAddress("0xYourContractAddress")

    // Create contract instance
    contract, err := bindings.NewYourContract(address, client)
    if err != nil {
        log.Fatal(err)
    }

    // Read from contract
    result, err := contract.SomeViewMethod(nil)
    if err != nil {
        log.Fatal(err)
    }

    // Write to contract (requires auth)
    // auth := bind.NewKeyedTransactor(privateKey)
    // tx, err := contract.SomeWriteMethod(auth, params...)
}
```

## Integration with Relayer

The relayer in this project can automatically interact with contracts using these bindings.

See `relayer/` for examples of how contracts are integrated with the event processing system.

## Note

- Generated files should not be manually edited
- Regenerate bindings when ABI changes
- Keep ABI files in sync with deployed contracts
