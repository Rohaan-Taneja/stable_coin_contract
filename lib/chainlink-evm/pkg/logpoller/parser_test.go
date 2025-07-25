package logpoller

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
	"github.com/smartcontractkit/chainlink-evm/pkg/types"
)

func assertArgs(t *testing.T, args *queryArgs, numVals int) {
	values, err := args.toArgs()

	assert.Len(t, values, numVals)
	assert.NoError(t, err)
}

func TestDSLParser(t *testing.T) {
	t.Parallel()

	t.Run("query with no filters no order and no limit", func(t *testing.T) {
		t.Parallel()

		parser := &pgDSLParser{}
		chainID := big.NewInt(1)
		expressions := []query.Expression{}
		limiter := query.LimitAndSort{}

		result, args, err := parser.buildQuery(chainID, expressions, limiter)

		require.NoError(t, err)
		assert.Equal(t, logsQuery(" WHERE evm_chain_id = :evm_chain_id ORDER BY "+defaultSort), result)

		assertArgs(t, args, 1)
	})

	t.Run("query with cursor and no order by", func(t *testing.T) {
		t.Parallel()

		parser := &pgDSLParser{}
		chainID := big.NewInt(1)
		expressions := []query.Expression{
			NewAddressFilter(common.HexToAddress("0x42")),
			NewEventSigFilter(common.HexToHash("0x21")),
			NewConfirmationsFilter(types.Finalized),
		}
		limiter := query.NewLimitAndSort(query.CursorLimit("10-5-0x42", query.CursorFollowing, 20))

		result, args, err := parser.buildQuery(chainID, expressions, limiter)
		expected := logsQuery(
			" WHERE evm_chain_id = :evm_chain_id " +
				"AND (address = :address_0 AND event_sig = :event_sig_0 " +
				"AND block_number <= " +
				"(SELECT finalized_block_number FROM evm.log_poller_blocks WHERE evm_chain_id = :evm_chain_id ORDER BY block_number DESC LIMIT 1)) " +
				"AND (block_number > :cursor_block_number OR (block_number = :cursor_block_number AND log_index > :cursor_log_index)) " +
				"ORDER BY block_number ASC, log_index ASC, tx_hash ASC " +
				"LIMIT 20")

		require.NoError(t, err)
		assert.Equal(t, expected, result)

		assertArgs(t, args, 5)
	})

	t.Run("query with limit and no order by", func(t *testing.T) {
		t.Parallel()

		parser := &pgDSLParser{}
		chainID := big.NewInt(1)
		expressions := []query.Expression{
			NewAddressFilter(common.HexToAddress("0x42")),
			NewEventSigFilter(common.HexToHash("0x21")),
		}
		limiter := query.NewLimitAndSort(query.CountLimit(20))

		result, args, err := parser.buildQuery(chainID, expressions, limiter)
		expected := logsQuery(
			" WHERE evm_chain_id = :evm_chain_id " +
				"AND (address = :address_0 AND event_sig = :event_sig_0) " +
				"ORDER BY " + defaultSort + " " +
				"LIMIT 20")

		require.NoError(t, err)
		assert.Equal(t, expected, result)

		assertArgs(t, args, 3)
	})

	t.Run("query with order by sequence no cursor no limit", func(t *testing.T) {
		t.Parallel()

		parser := &pgDSLParser{}
		chainID := big.NewInt(1)
		expressions := []query.Expression{}
		limiter := query.NewLimitAndSort(query.Limit{}, query.NewSortBySequence(query.Desc))

		result, args, err := parser.buildQuery(chainID, expressions, limiter)
		expected := logsQuery(
			" WHERE evm_chain_id = :evm_chain_id " +
				"ORDER BY block_number DESC, log_index DESC, tx_hash DESC")

		require.NoError(t, err)
		assert.Equal(t, expected, result)

		assertArgs(t, args, 1)
	})

	t.Run("query with multiple order by no limit", func(t *testing.T) {
		t.Parallel()

		parser := &pgDSLParser{}
		chainID := big.NewInt(1)
		expressions := []query.Expression{}
		limiter := query.NewLimitAndSort(query.Limit{}, query.NewSortByBlock(query.Asc), query.NewSortByTimestamp(query.Desc))

		result, args, err := parser.buildQuery(chainID, expressions, limiter)
		expected := logsQuery(
			" WHERE evm_chain_id = :evm_chain_id " +
				"ORDER BY block_number ASC, block_timestamp DESC")

		require.NoError(t, err)
		assert.Equal(t, expected, result)

		assertArgs(t, args, 1)
	})

	t.Run("basic query with default primitives no order by and cursor", func(t *testing.T) {
		t.Parallel()

		parser := &pgDSLParser{}
		chainID := big.NewInt(1)
		expressions := []query.Expression{
			query.Timestamp(10, primitives.Eq),
			query.TxHash(common.HexToHash("0x84").String()),
			query.Block("99", primitives.Neq),
			query.Confidence(primitives.Finalized),
		}
		limiter := query.NewLimitAndSort(query.CursorLimit("10-20-0x42", query.CursorPrevious, 20))

		result, args, err := parser.buildQuery(chainID, expressions, limiter)
		expected := logsQuery(
			" WHERE evm_chain_id = :evm_chain_id " +
				"AND (block_timestamp = :block_timestamp_0 " +
				"AND tx_hash = :tx_hash_0 " +
				"AND block_number != :block_number_0 " +
				"AND block_number <= " +
				"(SELECT finalized_block_number FROM evm.log_poller_blocks WHERE evm_chain_id = :evm_chain_id ORDER BY block_number DESC LIMIT 1)) " +
				"AND (block_number < :cursor_block_number OR (block_number = :cursor_block_number AND log_index < :cursor_log_index)) " +
				"ORDER BY block_number DESC, log_index DESC, tx_hash DESC LIMIT 20")

		require.NoError(t, err)
		assert.Equal(t, expected, result)

		assertArgs(t, args, 6)
	})

	t.Run("query for finality", func(t *testing.T) {
		t.Parallel()

		t.Run("finalized", func(t *testing.T) {
			parser := &pgDSLParser{}
			chainID := big.NewInt(1)

			expressions := []query.Expression{query.Confidence(primitives.Finalized)}
			limiter := query.LimitAndSort{}

			result, args, err := parser.buildQuery(chainID, expressions, limiter)
			expected := logsQuery(
				" WHERE evm_chain_id = :evm_chain_id " +
					"AND block_number <= (SELECT finalized_block_number FROM evm.log_poller_blocks WHERE evm_chain_id = :evm_chain_id ORDER BY block_number DESC LIMIT 1) ORDER BY " + defaultSort)

			require.NoError(t, err)
			assert.Equal(t, expected, result)

			assertArgs(t, args, 1)
		})

		t.Run("safe", func(t *testing.T) {
			parser := &pgDSLParser{}
			chainID := big.NewInt(1)

			expressions := []query.Expression{query.Confidence(primitives.Safe)}
			limiter := query.LimitAndSort{}

			result, args, err := parser.buildQuery(chainID, expressions, limiter)
			expected := logsQuery(
				" WHERE evm_chain_id = :evm_chain_id " +
					"AND block_number <= (SELECT safe_block_number FROM evm.log_poller_blocks WHERE evm_chain_id = :evm_chain_id ORDER BY block_number DESC LIMIT 1) ORDER BY " + defaultSort)

			require.NoError(t, err)
			assert.Equal(t, expected, result)

			assertArgs(t, args, 1)
		})

		t.Run("unconfirmed", func(t *testing.T) {
			parser := &pgDSLParser{}
			chainID := big.NewInt(1)

			expressions := []query.Expression{query.Confidence(primitives.Unconfirmed)}
			limiter := query.LimitAndSort{}

			result, args, err := parser.buildQuery(chainID, expressions, limiter)
			expected := logsQuery(
				" WHERE evm_chain_id = :evm_chain_id " +
					"AND block_number <= (SELECT greatest(block_number - :confs_0, 0) FROM evm.log_poller_blocks WHERE evm_chain_id = :evm_chain_id ORDER BY block_number DESC LIMIT 1) ORDER BY " + defaultSort)

			require.NoError(t, err)
			assert.Equal(t, expected, result)

			assertArgs(t, args, 2)
		})

		t.Run("exact confirmations", func(t *testing.T) {
			parser := &pgDSLParser{}
			chainID := big.NewInt(1)

			expressions := []query.Expression{NewConfirmationsFilter(25)}
			limiter := query.LimitAndSort{}

			result, args, err := parser.buildQuery(chainID, expressions, limiter)
			expected := logsQuery(
				" WHERE evm_chain_id = :evm_chain_id " +
					"AND block_number <= (SELECT greatest(block_number - :confs_0, 0) FROM evm.log_poller_blocks WHERE evm_chain_id = :evm_chain_id ORDER BY block_number DESC LIMIT 1) ORDER BY " + defaultSort)

			require.NoError(t, err)
			assert.Equal(t, expected, result)

			confirmations, ok := args.args["confs_0"]

			require.True(t, ok)
			require.Equal(t, uint64(25), confirmations)

			assertArgs(t, args, 2)
		})
	})

	t.Run("query for event by word", func(t *testing.T) {
		t.Parallel()

		wordFilter := NewEventByWordFilter(8, []HashedValueComparator{
			{Values: []common.Hash{common.HexToHash("0x1"), common.HexToHash("0x2")}, Operator: primitives.Gt},
		})

		parser := &pgDSLParser{}
		chainID := big.NewInt(1)
		expressions := []query.Expression{wordFilter}
		limiter := query.LimitAndSort{}

		result, args, err := parser.buildQuery(chainID, expressions, limiter)
		expected := logsQuery(
			" WHERE evm_chain_id = :evm_chain_id " +
				"AND substring(data from 32*8+1 for 32) > ANY(:word_value_0) ORDER BY " + defaultSort)

		require.NoError(t, err)
		assert.Equal(t, expected, result)

		values, err := args.toArgs()
		require.NoError(t, err)
		require.Len(t, values, 2)
		// HashedValueComparator values should be concatenated into single slice
		require.Len(t, values["word_value_0"], 2)
	})

	t.Run("query for event topic", func(t *testing.T) {
		t.Parallel()

		topicFilter := NewEventByTopicFilter(2, []HashedValueComparator{
			{Values: []common.Hash{common.HexToHash("a")}, Operator: primitives.Gt},
			{Values: []common.Hash{common.HexToHash("b"), common.HexToHash("c")}, Operator: primitives.Lt},
		})

		parser := &pgDSLParser{}
		chainID := big.NewInt(1)
		expressions := []query.Expression{topicFilter}
		limiter := query.LimitAndSort{}

		result, args, err := parser.buildQuery(chainID, expressions, limiter)
		expected := logsQuery(
			" WHERE evm_chain_id = :evm_chain_id " +
				"AND topics[3] > :topic_value_0 AND topics[3] < ANY(:topic_value_1) ORDER BY " + defaultSort)

		require.NoError(t, err)
		assert.Equal(t, expected, result)

		assertArgs(t, args, 3)
	})

	// nested query -> a & (b || c)
	t.Run("nested query", func(t *testing.T) {
		t.Parallel()

		parser := &pgDSLParser{}
		chainID := big.NewInt(1)

		expressions := []query.Expression{
			{BoolExpression: query.BoolExpression{
				Expressions: []query.Expression{
					query.Timestamp(10, primitives.Gte),
					{BoolExpression: query.BoolExpression{
						Expressions: []query.Expression{
							query.TxHash(common.HexToHash("0x84").Hex()),
							query.Confidence(primitives.Unconfirmed),
						},
						BoolOperator: query.OR,
					}},
				},
				BoolOperator: query.AND,
			}},
		}
		limiter := query.LimitAndSort{}

		result, args, err := parser.buildQuery(chainID, expressions, limiter)
		expected := logsQuery(
			" WHERE evm_chain_id = :evm_chain_id " +
				"AND (block_timestamp >= :block_timestamp_0 " +
				"AND (tx_hash = :tx_hash_0 " +
				"OR block_number <= (SELECT greatest(block_number - :confs_0, 0) FROM evm.log_poller_blocks WHERE evm_chain_id = :evm_chain_id ORDER BY block_number DESC LIMIT 1))) ORDER BY " + defaultSort)

		require.NoError(t, err)
		assert.Equal(t, expected, result)

		assertArgs(t, args, 4)
	})

	// deep nested query -> a & (b || (c & d))
	t.Run("nested query deep", func(t *testing.T) {
		t.Parallel()

		wordFilter := NewEventByWordFilter(8, []HashedValueComparator{
			{Values: []common.Hash{common.HexToHash("a")}, Operator: primitives.Gt},
			{Values: []common.Hash{common.HexToHash("b"), common.HexToHash("c")}, Operator: primitives.Lte},
		})

		parser := &pgDSLParser{}
		chainID := big.NewInt(1)

		expressions := []query.Expression{
			{BoolExpression: query.BoolExpression{
				Expressions: []query.Expression{
					query.Timestamp(10, primitives.Eq),
					{BoolExpression: query.BoolExpression{
						Expressions: []query.Expression{
							query.TxHash(common.HexToHash("0x84").Hex()),
							{BoolExpression: query.BoolExpression{
								Expressions: []query.Expression{
									query.Confidence(primitives.Unconfirmed),
									wordFilter,
								},
								BoolOperator: query.AND,
							}},
						},
						BoolOperator: query.OR,
					}},
				},
				BoolOperator: query.AND,
			}},
		}
		limiter := query.LimitAndSort{}

		result, args, err := parser.buildQuery(chainID, expressions, limiter)
		expected := logsQuery(
			" WHERE evm_chain_id = :evm_chain_id " +
				"AND (block_timestamp = :block_timestamp_0 " +
				"AND (tx_hash = :tx_hash_0 " +
				"OR (block_number <= (SELECT greatest(block_number - :confs_0, 0) FROM evm.log_poller_blocks WHERE evm_chain_id = :evm_chain_id ORDER BY block_number DESC LIMIT 1) " +
				"AND substring(data from 32*8+1 for 32) > :word_value_0 " +
				"AND substring(data from 32*8+1 for 32) <= ANY(:word_value_1)))) ORDER BY " + defaultSort)

		require.NoError(t, err)
		assert.Equal(t, expected, result)

		values, err := args.toArgs()
		require.NoError(t, err)
		require.Len(t, values, 6)
		// unwraps slice of len 1
		require.IsType(t, []uint8{}, values["word_value_0"])
		// HashedValueComparator values should be concatenated into single slice
		require.IsType(t, [][]uint8{}, values["word_value_1"])
		require.Len(t, values["word_value_1"], 2)
	})
}
