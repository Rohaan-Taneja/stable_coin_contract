package rollups

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"

	"github.com/smartcontractkit/chainlink-evm/pkg/assets"
	"github.com/smartcontractkit/chainlink-evm/pkg/client"
	evmconfig "github.com/smartcontractkit/chainlink-evm/pkg/config"
	"github.com/smartcontractkit/chainlink-evm/pkg/config/chaintype"
	"github.com/smartcontractkit/chainlink-evm/pkg/config/toml"
)

// Reads L2-specific precompiles and caches the l1GasPrice set by the L2.
//
//nolint:unused // backported from CCIP
type optimismL1Oracle struct {
	services.StateMachine
	client     l1OracleClient
	pollPeriod time.Duration
	logger     logger.SugaredLogger

	daOracleAddress common.Address
	l1GasPriceMu    sync.RWMutex
	l1GasPrice      priceEntry
	isEcotone       bool
	isFjord         bool
	upgradeCheckTs  time.Time

	chInitialised chan struct{}
	chStop        services.StopChan
	chDone        chan struct{}

	getL1FeeMethodAbi         abi.ABI
	l1BaseFeeCalldata         []byte
	baseFeeScalarCalldata     []byte
	blobBaseFeeCalldata       []byte
	blobBaseFeeScalarCalldata []byte
	decimalsCalldata          []byte
	tokenRatioCalldata        []byte
	isEcotoneCalldata         []byte
	isEcotoneMethodAbi        abi.ABI
	isFjordCalldata           []byte
	isFjordMethodAbi          abi.ABI
}

const (
	// upgradePollingPeriod is the interval to poll if chain has been upgraded
	upgradePollingPeriod = 4 * time.Hour
	// isEcotone fetches if the OP Stack GasPriceOracle contract has upgraded to Ecotone
	isEcotoneMethod = "isEcotone"
	// isFjord fetches if the OP Stack GasPriceOracle contract has upgraded to Fjord
	isFjordMethod = "isFjord"
	// getL1Fee fetches the l1 fee for given tx bytes
	// getL1Fee is a hex encoded call to:
	// `function getL1Fee(bytes) external view returns (uint256);`
	getL1FeeMethod = "getL1Fee"
	// l1BaseFee fetches the l1 base fee set in the OP Stack GasPriceOracle contract
	// l1BaseFee is a hex encoded call to:
	// `function l1BaseFee() external view returns (uint256);`
	l1BaseFeeMethod = "l1BaseFee"
	// baseFeeScalar fetches the l1 base fee scalar for gas price calculation
	// baseFeeScalar is a hex encoded call to:
	// `function baseFeeScalar() public view returns (uint32);`
	baseFeeScalarMethod = "baseFeeScalar"
	// blobBaseFee fetches the l1 blob base fee for gas price calculation
	// blobBaseFee is a hex encoded call to:
	// `function blobBaseFee() public view returns (uint256);`
	blobBaseFeeMethod = "blobBaseFee"
	// blobBaseFeeScalar fetches the l1 blob base fee scalar for gas price calculation
	// blobBaseFeeScalar is a hex encoded call to:
	// `function blobBaseFeeScalar() public view returns (uint32);`
	blobBaseFeeScalarMethod = "blobBaseFeeScalar"
	// decimals fetches the number of decimals used in the scalar for gas price calculation
	// decimals is a hex encoded call to:
	// `function decimals() public pure returns (uint256);`
	decimalsMethod = "decimals"
)

func NewOpStackL1GasOracle(lggr logger.Logger, ethClient l1OracleClient, chainType chaintype.ChainType, daOracle evmconfig.DAOracle) (*optimismL1Oracle, error) {
	if daOracle.OracleType() == nil {
		return nil, errors.New("OracleType is required but was nil")
	}
	if *daOracle.OracleType() != toml.DAOracleOPStack {
		return nil, fmt.Errorf("expected %s oracle type, got %s", toml.DAOracleOPStack, *daOracle.OracleType())
	}
	if daOracle.CustomGasPriceCalldata() != nil && *daOracle.CustomGasPriceCalldata() != "" {
		lggr.Warnf("CustomGasPriceCalldata is set but will be ignored for OPStack DA oracle")
	}
	if daOracle.OracleAddress() == nil || *daOracle.OracleAddress() == "" {
		return nil, errors.New("OracleAddress is required but was nil or empty")
	}
	oracleAddress := *daOracle.OracleAddress()

	getL1FeeMethodAbi, err := abi.JSON(strings.NewReader(GetL1FeeAbiString))
	if err != nil {
		return nil, fmt.Errorf("failed to parse L1 gas cost method ABI for chain: %s", chainType)
	}

	// encode calldata for each method; these calldata will remain the same for each call, we can encode them just once
	// Encode calldata for l1BaseFee method
	l1BaseFeeCalldata, _, err := encodeCalldata(L1BaseFeeAbiString, l1BaseFeeMethod)
	if err != nil {
		return nil, fmt.Errorf("failed to parse GasPriceOracle %s() calldata for chain: %s; %w", l1BaseFeeMethod, chainType, err)
	}

	// Encode calldata for isEcotone method
	isEcotoneCalldata, isEcotoneMethodAbi, err := encodeCalldata(OPIsEcotoneAbiString, isEcotoneMethod)
	if err != nil {
		return nil, fmt.Errorf("failed to parse GasPriceOracle %s() calldata for chain: %s; %w", isEcotoneMethod, chainType, err)
	}

	// Encode calldata for isFjord method
	isFjordCalldata, isFjordMethodAbi, err := encodeCalldata(OPIsFjordAbiString, isFjordMethod)
	if err != nil {
		return nil, fmt.Errorf("failed to parse GasPriceOracle %s() calldata for chain: %s; %w", isFjordMethod, chainType, err)
	}

	// Encode calldata for baseFeeScalar method
	baseFeeScalarCalldata, _, err := encodeCalldata(OPBaseFeeScalarAbiString, baseFeeScalarMethod)
	if err != nil {
		return nil, fmt.Errorf("failed to parse GasPriceOracle %s() calldata for chain: %s; %w", baseFeeScalarMethod, chainType, err)
	}

	// Encode calldata for blobBaseFee method
	blobBaseFeeCalldata, _, err := encodeCalldata(OPBlobBaseFeeAbiString, blobBaseFeeMethod)
	if err != nil {
		return nil, fmt.Errorf("failed to parse GasPriceOracle %s() calldata for chain: %s; %w", blobBaseFeeMethod, chainType, err)
	}

	// Encode calldata for blobBaseFeeScalar method
	blobBaseFeeScalarCalldata, _, err := encodeCalldata(OPBlobBaseFeeScalarAbiString, blobBaseFeeScalarMethod)
	if err != nil {
		return nil, fmt.Errorf("failed to parse GasPriceOracle %s() calldata for chain: %s; %w", blobBaseFeeScalarMethod, chainType, err)
	}

	// Encode calldata for decimals method
	decimalsCalldata, _, err := encodeCalldata(OPDecimalsAbiString, decimalsMethod)
	if err != nil {
		return nil, fmt.Errorf("failed to parse GasPriceOracle %s() calldata for chain: %s; %w", decimalsMethod, chainType, err)
	}

	return &optimismL1Oracle{
		client:     ethClient,
		pollPeriod: PollPeriod,
		logger:     logger.Sugared(logger.Named(lggr, fmt.Sprintf("L1GasOracle(%s)", chainType))),

		daOracleAddress: oracleAddress.Address(),
		isEcotone:       false,
		isFjord:         false,
		upgradeCheckTs:  time.Time{},

		chInitialised: make(chan struct{}),
		chStop:        make(chan struct{}),
		chDone:        make(chan struct{}),

		getL1FeeMethodAbi:         getL1FeeMethodAbi,
		l1BaseFeeCalldata:         l1BaseFeeCalldata,
		baseFeeScalarCalldata:     baseFeeScalarCalldata,
		blobBaseFeeCalldata:       blobBaseFeeCalldata,
		blobBaseFeeScalarCalldata: blobBaseFeeScalarCalldata,
		decimalsCalldata:          decimalsCalldata,
		isEcotoneCalldata:         isEcotoneCalldata,
		isEcotoneMethodAbi:        isEcotoneMethodAbi,
		isFjordCalldata:           isFjordCalldata,
		isFjordMethodAbi:          isFjordMethodAbi,
	}, nil
}

func (o *optimismL1Oracle) Name() string {
	return o.logger.Name()
}

func (o *optimismL1Oracle) Start(ctx context.Context) error {
	return o.StartOnce(o.Name(), func() error {
		go o.run()
		<-o.chInitialised
		return nil
	})
}
func (o *optimismL1Oracle) Close() error {
	return o.StopOnce(o.Name(), func() error {
		close(o.chStop)
		<-o.chDone
		return nil
	})
}

func (o *optimismL1Oracle) HealthReport() map[string]error {
	return map[string]error{o.Name(): o.Healthy()}
}

func (o *optimismL1Oracle) run() {
	defer close(o.chDone)

	o.refresh()
	close(o.chInitialised)

	t := services.TickerConfig{
		Initial:   o.pollPeriod,
		JitterPct: services.DefaultJitter,
	}.NewTicker(o.pollPeriod)
	defer t.Stop()

	for {
		select {
		case <-o.chStop:
			return
		case <-t.C:
			o.refresh()
		}
	}
}

func (o *optimismL1Oracle) refresh() {
	err := o.refreshWithError()
	if err != nil {
		o.logger.Criticalw("Failed to refresh gas price", "err", err)
		o.SvcErrBuffer.Append(err)
	}
}

func (o *optimismL1Oracle) refreshWithError() error {
	ctx, cancel := o.chStop.CtxWithTimeout(client.QueryTimeout)
	defer cancel()

	price, err := o.GetDAGasPrice(ctx)
	if err != nil {
		return err
	}

	o.l1GasPriceMu.Lock()
	defer o.l1GasPriceMu.Unlock()
	o.l1GasPrice = priceEntry{price: assets.NewWei(price), timestamp: time.Now()}
	return nil
}

func (o *optimismL1Oracle) GasPrice(_ context.Context) (l1GasPrice *assets.Wei, err error) {
	var timestamp time.Time
	ok := o.IfStarted(func() {
		o.l1GasPriceMu.RLock()
		l1GasPrice = o.l1GasPrice.price
		timestamp = o.l1GasPrice.timestamp
		o.l1GasPriceMu.RUnlock()
	})
	if !ok {
		return l1GasPrice, errors.New("L1GasOracle is not started; cannot estimate gas")
	}
	if l1GasPrice == nil {
		return l1GasPrice, errors.New("failed to get l1 gas price; gas price not set")
	}
	// Validate the price has been updated within the pollPeriod * 2
	// Allowing double the poll period before declaring the price stale to give ample time for the refresh to process
	if time.Since(timestamp) > o.pollPeriod*2 {
		return l1GasPrice, errors.New("gas price is stale")
	}
	return
}

func (o *optimismL1Oracle) GetDAGasPrice(ctx context.Context) (*big.Int, error) {
	err := o.checkForUpgrade(ctx)
	if err != nil {
		return nil, err
	}
	if o.isFjord || o.isEcotone {
		return o.getEcotoneFjordGasPrice(ctx)
	}

	return o.getV1GasPrice(ctx)
}

// Checks oracle flags for Ecotone and Fjord upgrades
func (o *optimismL1Oracle) checkForUpgrade(ctx context.Context) error {
	// if chain is already Fjord (the latest upgrade), NOOP
	// need to continue to check if not on latest upgrade
	if o.isFjord {
		return nil
	}
	// if time since last check has not exceeded polling period, NOOP
	if time.Since(o.upgradeCheckTs) < upgradePollingPeriod {
		return nil
	}

	o.upgradeCheckTs = time.Now()
	rpcBatchCalls := []rpc.BatchElem{
		{
			Method: "eth_call",
			Args: []any{
				map[string]interface{}{
					"from": common.Address{},
					"to":   o.daOracleAddress.String(),
					"data": hexutil.Bytes(o.isFjordCalldata),
				},
				"latest",
			},
			Result: new(string),
		},
		{
			Method: "eth_call",
			Args: []any{
				map[string]interface{}{
					"from": common.Address{},
					"to":   o.daOracleAddress.String(),
					"data": hexutil.Bytes(o.isEcotoneCalldata),
				},
				"latest",
			},
			Result: new(string),
		},
	}
	err := o.client.BatchCallContext(ctx, rpcBatchCalls)
	if err != nil {
		return fmt.Errorf("check upgrade batch call failed: %w", err)
	}
	// These calls are expected to revert if chain has not upgraded. Ignore non-nil Error field.
	if rpcBatchCalls[0].Error == nil {
		result := *(rpcBatchCalls[0].Result.(*string))
		if b, decodeErr := hexutil.Decode(result); decodeErr == nil {
			if res, unpackErr := o.isFjordMethodAbi.Unpack(isFjordMethod, b); unpackErr == nil {
				o.isFjord = res[0].(bool)
			} else {
				o.logger.Errorw("failed to unpack results", "method", isFjordMethod, "hex", result, "error", unpackErr)
			}
		} else {
			o.logger.Errorw("failed to decode bytes", "method", isFjordMethod, "hex", result, "error", decodeErr)
		}
	}
	if rpcBatchCalls[1].Error == nil {
		result := *(rpcBatchCalls[1].Result.(*string))
		if b, decodeErr := hexutil.Decode(result); decodeErr == nil {
			if res, unpackErr := o.isEcotoneMethodAbi.Unpack(isEcotoneMethod, b); unpackErr == nil {
				o.isEcotone = res[0].(bool)
			} else {
				o.logger.Errorw("failed to unpack results", "method", isEcotoneMethod, "hex", result, "error", unpackErr)
			}
		} else {
			o.logger.Errorw("failed to decode bytes", "method", isEcotoneMethod, "hex", result, "error", decodeErr)
		}
	}
	return nil
}

func (o *optimismL1Oracle) getV1GasPrice(ctx context.Context) (*big.Int, error) {
	b, err := o.client.CallContract(ctx, ethereum.CallMsg{
		To:   &o.daOracleAddress,
		Data: o.l1BaseFeeCalldata,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("l1BaseFee() call failed: %w", err)
	}

	if len(b) != 32 {
		return nil, fmt.Errorf("l1BaseFee() return data length (%d) different than expected (%d)", len(b), 32)
	}
	return new(big.Int).SetBytes(b), nil
}

// Returns the scaled gas price using baseFeeScalar, l1BaseFee, blobBaseFeeScalar, and blobBaseFee fields from the oracle
// Confirmed the same calculation is used to determine gas price for both Ecotone and Fjord
func (o *optimismL1Oracle) getEcotoneFjordGasPrice(ctx context.Context) (*big.Int, error) {
	rpcBatchCalls := []rpc.BatchElem{
		{
			Method: "eth_call",
			Args: []any{
				map[string]interface{}{
					"from": common.Address{},
					"to":   o.daOracleAddress.String(),
					"data": hexutil.Bytes(o.l1BaseFeeCalldata),
				},
				"latest",
			},
			Result: new(string),
		},
		{
			Method: "eth_call",
			Args: []any{
				map[string]interface{}{
					"from": common.Address{},
					"to":   o.daOracleAddress.String(),
					"data": hexutil.Bytes(o.baseFeeScalarCalldata),
				},
				"latest",
			},
			Result: new(string),
		},
		{
			Method: "eth_call",
			Args: []any{
				map[string]interface{}{
					"from": common.Address{},
					"to":   o.daOracleAddress.String(),
					"data": hexutil.Bytes(o.blobBaseFeeCalldata),
				},
				"latest",
			},
			Result: new(string),
		},
		{
			Method: "eth_call",
			Args: []any{
				map[string]interface{}{
					"from": common.Address{},
					"to":   o.daOracleAddress.String(),
					"data": hexutil.Bytes(o.blobBaseFeeScalarCalldata),
				},
				"latest",
			},
			Result: new(string),
		},
		{
			Method: "eth_call",
			Args: []any{
				map[string]interface{}{
					"from": common.Address{},
					"to":   o.daOracleAddress.String(),
					"data": hexutil.Bytes(o.decimalsCalldata),
				},
				"latest",
			},
			Result: new(string),
		},
	}

	err := o.client.BatchCallContext(ctx, rpcBatchCalls)
	if err != nil {
		return nil, fmt.Errorf("fetch gas price parameters batch call failed: %w", err)
	}
	if rpcBatchCalls[0].Error != nil {
		return nil, fmt.Errorf("%s call failed in a batch: %w", l1BaseFeeMethod, err)
	}
	if rpcBatchCalls[1].Error != nil {
		return nil, fmt.Errorf("%s call failed in a batch: %w", baseFeeScalarMethod, err)
	}
	if rpcBatchCalls[2].Error != nil {
		return nil, fmt.Errorf("%s call failed in a batch: %w", blobBaseFeeMethod, err)
	}
	if rpcBatchCalls[3].Error != nil {
		return nil, fmt.Errorf("%s call failed in a batch: %w", blobBaseFeeScalarMethod, err)
	}
	if rpcBatchCalls[4].Error != nil {
		return nil, fmt.Errorf("%s call failed in a batch: %w", decimalsMethod, err)
	}

	// Extract values from responses
	l1BaseFeeResult := *(rpcBatchCalls[0].Result.(*string))
	baseFeeScalarResult := *(rpcBatchCalls[1].Result.(*string))
	blobBaseFeeResult := *(rpcBatchCalls[2].Result.(*string))
	blobBaseFeeScalarResult := *(rpcBatchCalls[3].Result.(*string))
	decimalsResult := *(rpcBatchCalls[4].Result.(*string))

	// Decode the responses into bytes
	l1BaseFeeBytes, err := hexutil.Decode(l1BaseFeeResult)
	if err != nil {
		return nil, fmt.Errorf("failed to decode %s rpc result: %w", l1BaseFeeMethod, err)
	}
	baseFeeScalarBytes, err := hexutil.Decode(baseFeeScalarResult)
	if err != nil {
		return nil, fmt.Errorf("failed to decode %s rpc result: %w", baseFeeScalarMethod, err)
	}
	blobBaseFeeBytes, err := hexutil.Decode(blobBaseFeeResult)
	if err != nil {
		return nil, fmt.Errorf("failed to decode %s rpc result: %w", blobBaseFeeMethod, err)
	}
	blobBaseFeeScalarBytes, err := hexutil.Decode(blobBaseFeeScalarResult)
	if err != nil {
		return nil, fmt.Errorf("failed to decode %s rpc result: %w", blobBaseFeeScalarMethod, err)
	}
	decimalsBytes, err := hexutil.Decode(decimalsResult)
	if err != nil {
		return nil, fmt.Errorf("failed to decode %s rpc result: %w", decimalsMethod, err)
	}

	// Convert bytes to big int for calculations
	l1BaseFee := new(big.Int).SetBytes(l1BaseFeeBytes)
	baseFeeScalar := new(big.Int).SetBytes(baseFeeScalarBytes)
	blobBaseFee := new(big.Int).SetBytes(blobBaseFeeBytes)
	blobBaseFeeScalar := new(big.Int).SetBytes(blobBaseFeeScalarBytes)
	decimals := new(big.Int).SetBytes(decimalsBytes)

	o.logger.Debugw("gas price parameters", "l1BaseFee", l1BaseFee, "baseFeeScalar", baseFeeScalar, "blobBaseFee", blobBaseFee, "blobBaseFeeScalar", blobBaseFeeScalar, "decimals", decimals)

	// Scaled gas price = baseFee * 16 * baseFeeScalar + blobBaseFee * blobBaseFeeScalar
	scaledBaseFee := new(big.Int).Mul(l1BaseFee, baseFeeScalar)
	scaledBaseFee = new(big.Int).Mul(scaledBaseFee, big.NewInt(16))
	scaledBlobBaseFee := new(big.Int).Mul(blobBaseFee, blobBaseFeeScalar)
	scaledGasPrice := new(big.Int).Add(scaledBaseFee, scaledBlobBaseFee)

	// Gas price = scaled gas price / (16 * 10 ^ decimals)
	// This formula is extracted from the gas cost methods in the precompile contract
	// Note: The Fjord calculation in the contract uses estimated size instead of gas used which is why we have to scale down by (16 * 10 ^ decimals) as well
	// Ecotone: https://github.com/ethereum-optimism/optimism/blob/71b93116738ee98c9f8713b1a5dfe626ce06c1b2/packages/contracts-bedrock/src/L2/GasPriceOracle.sol#L192
	// Fjord: https://github.com/ethereum-optimism/optimism/blob/71b93116738ee98c9f8713b1a5dfe626ce06c1b2/packages/contracts-bedrock/src/L2/GasPriceOracle.sol#L229-L230
	scale := new(big.Int).Exp(big.NewInt(10), decimals, nil)
	scale = new(big.Int).Mul(scale, big.NewInt(16))

	return new(big.Int).Div(scaledGasPrice, scale), nil
}

func encodeCalldata(abiString, methodName string) ([]byte, abi.ABI, error) {
	methodAbi, err := abi.JSON(strings.NewReader(abiString))
	if err != nil {
		return nil, abi.ABI{}, fmt.Errorf("failed to parse ABI: %w", err)
	}
	calldata, err := methodAbi.Pack(methodName)
	if err != nil {
		return nil, abi.ABI{}, fmt.Errorf("failed to pack calldata for %s: %w", methodName, err)
	}
	return calldata, methodAbi, nil
}
