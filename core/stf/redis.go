package stf

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-redis/redis/v8"
	"strconv"
)

func SetAddress(rdb *redis.Client, address common.Address) error {
	addressStr := address.Hex()
	err := rdb.Set(context.Background(), addressStr, "0", 0).Err()
	return err
}

func GetAddressValue(rdb *redis.Client, address common.Address) (string, error) {
	addressStr := address.Hex()
	value, err := rdb.Get(context.Background(), addressStr).Result()
	return value, err
}

func CheckKeyExists(rdb *redis.Client, address common.Address) (bool, error) {
	addressStr := address.Hex()
	keyExists, err := rdb.Exists(context.Background(), addressStr).Result()
	return keyExists == 1, err
}

func GetProcessedBlocks(ctx context.Context, rdb *redis.Client) ([]uint64, error) {
	blockNumStrings, err := rdb.SMembers(ctx, "processedBlocks").Result()
	if err != nil {
		return nil, err
	}

	blockNums := make([]uint64, len(blockNumStrings))
	for i, numStr := range blockNumStrings {
		blockNum, err := strconv.ParseUint(numStr, 10, 64)
		if err != nil {
			return nil, err
		}
		blockNums[i] = blockNum
	}

	return blockNums, nil
}

func UpdateLastTenBlocks(ctx context.Context, rdb *redis.Client, blockNum uint64) error {
	rdb.LPush(ctx, "lastTenBlocks", blockNum)
	rdb.LTrim(ctx, "lastTenBlocks", 0, 99)
	return nil
}

func GetLastTenBlocks(ctx context.Context, rdb *redis.Client) ([]uint64, error) {
	blockNumStrings, err := rdb.LRange(ctx, "lastTenBlocks", 0, -1).Result()
	if err != nil {
		return nil, err
	}

	blockNums := make([]uint64, len(blockNumStrings))
	for i, numStr := range blockNumStrings {
		blockNum, err := strconv.ParseUint(numStr, 10, 64)
		if err != nil {
			return nil, err
		}
		blockNums[i] = blockNum
	}

	return blockNums, nil
}

func AddProcessedBlock(ctx context.Context, rdb *redis.Client, blockNum uint64) error {
	rdb.SAdd(ctx, "processedBlocks", blockNum)
	return nil
}
