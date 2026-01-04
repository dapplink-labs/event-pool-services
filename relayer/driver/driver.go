package driver

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"time"

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"

	// TODO: Import your generated contract bindings here
	// "github.com/multimarket-labs/event-pod-services/bindings"
	"github.com/multimarket-labs/event-pod-services/metrics"
	"github.com/multimarket-labs/event-pod-services/relayer/txmgr"
)

var (
	errMaxPriorityFeePerGasNotFound = errors.New("Method eth_maxPriorityFeePerGas not found")
	FallbackGasTipCap               = big.NewInt(1500000000)
)

type DriverEngineConfig struct {
	ChainClient               *ethclient.Client
	ChainId                   *big.Int
	ContractAddress           common.Address // Your contract address
	CallerAddress             common.Address
	PrivateKey                *ecdsa.PrivateKey
	NumConfirmations          uint64
	SafeAbortNonceTooLowCount uint64
}

type DriverEngine struct {
	Ctx                 context.Context
	Cfg                 *DriverEngineConfig
	// TODO: Add your contract instance here
	// YourContract        *bindings.YourContract
	RawContract         *bind.BoundContract
	ContractAbi         *abi.ABI
	TxMgr               txmgr.TxManager
	Metrics             *metrics.PhoenixMetrics
	cancel              func()
}

func NewDriverEngine(ctx context.Context, phoenixMetrics *metrics.PhoenixMetrics, cfg *DriverEngineConfig) (*DriverEngine, error) {
	_, cancel := context.WithTimeout(ctx, time.Second*15)
	defer cancel()

	// TODO: Initialize your contract binding
	// Example:
	// yourContract, err := bindings.NewYourContract(cfg.ContractAddress, cfg.ChainClient)
	// if err != nil {
	//     log.Error("failed to create contract instance", "err", err)
	//     return nil, err
	// }

	// TODO: Parse your contract ABI
	// parsed, err := abi.JSON(strings.NewReader(bindings.YourContractMetaData.ABI))
	// if err != nil {
	//     log.Error("failed to parse abi", "err", err)
	//     return nil, err
	// }

	// TODO: Get your contract ABI
	// contractAbi, err := bindings.YourContractMetaData.GetAbi()
	// if err != nil {
	//     log.Error("failed to get contract abi", "err", err)
	//     return nil, err
	// }

	// TODO: Create bound contract
	// rawContract := bind.NewBoundContract(cfg.ContractAddress, parsed, cfg.ChainClient, cfg.ChainClient, cfg.ChainClient)

	txManagerConfig := txmgr.Config{
		ResubmissionTimeout:       time.Second * 5,
		ReceiptQueryInterval:      time.Second,
		NumConfirmations:          cfg.NumConfirmations,
		SafeAbortNonceTooLowCount: cfg.SafeAbortNonceTooLowCount,
	}

	txManager := txmgr.NewSimpleTxManager(txManagerConfig, cfg.ChainClient)

	return &DriverEngine{
		Ctx:         ctx,
		Cfg:         cfg,
		// YourContract:        yourContract,
		// RawContract:         rawContract,
		// ContractAbi:         contractAbi,
		TxMgr:       txManager,
		Metrics:     phoenixMetrics,
		cancel:      cancel,
	}, nil
}

func (de *DriverEngine) UpdateGasPrice(ctx context.Context, tx *types.Transaction) (*types.Transaction, error) {
	var opts *bind.TransactOpts
	var err error
	opts, err = bind.NewKeyedTransactorWithChainID(de.Cfg.PrivateKey, de.Cfg.ChainId)
	if err != nil {
		log.Error("new keyed transactor with chain id fail", "err", err)
		return nil, err
	}

	opts.Context = ctx
	opts.Nonce = new(big.Int).SetUint64(tx.Nonce())
	opts.NoSend = true

	finalGasTipCap, finalGasFeeCap, err := de.SuggestGasPriceCaps(ctx)
	if err != nil {
		log.Warn("failed to get suggested gas price", "err", err)
		return nil, err
	}

	if tx.GasTipCap().Cmp(finalGasTipCap) > 0 {
		finalGasTipCap = tx.GasTipCap()
	}

	if tx.GasFeeCap().Cmp(finalGasFeeCap) > 0 {
		finalGasFeeCap = tx.GasFeeCap()
	}

	opts.GasTipCap = finalGasTipCap
	opts.GasFeeCap = finalGasFeeCap

	return types.NewTx(&types.DynamicFeeTx{
		ChainID:   de.Cfg.ChainId,
		Nonce:     opts.Nonce.Uint64(),
		GasTipCap: opts.GasTipCap,
		GasFeeCap: opts.GasFeeCap,
		Gas:       tx.Gas(),
		To:        tx.To(),
		Value:     tx.Value(),
		Data:      tx.Data(),
	}), nil
}

func (de *DriverEngine) SuggestGasPriceCaps(ctx context.Context) (*big.Int, *big.Int, error) {
	gasTipCap, err := de.Cfg.ChainClient.SuggestGasTipCap(ctx)
	if err != nil {
		if errors.Is(err, errMaxPriorityFeePerGasNotFound) {
			gasTipCap = FallbackGasTipCap
		} else {
			return nil, nil, err
		}
	}

	head, err := de.Cfg.ChainClient.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	gasFeeCap := txmgr.CalcGasFeeCap(head.BaseFee, gasTipCap)

	return gasTipCap, gasFeeCap, nil
}

// TODO: Add your contract-specific methods here
// Example:
//
// func (de *DriverEngine) CallContractMethod(ctx context.Context, param1 string) error {
//     opts, err := bind.NewKeyedTransactorWithChainID(de.Cfg.PrivateKey, de.Cfg.ChainId)
//     if err != nil {
//         return err
//     }
//
//     tx, err := de.YourContract.YourMethod(opts, param1)
//     if err != nil {
//         return err
//     }
//
//     receipt, err := de.TxMgr.Send(ctx, txmgr.TxCandidate{
//         To:       &de.Cfg.ContractAddress,
//         TxData:   tx.Data(),
//         GasLimit: tx.Gas(),
//         Value:    tx.Value(),
//     })
//
//     if err != nil {
//         return err
//     }
//
//     log.Info("transaction sent", "hash", receipt.TxHash.Hex())
//     return nil
// }
