package task

import "math/big"

type Task struct {
	Idx int64
	Epochs []*GlobalEpoch
	OriginatorID int64
}

type GlobalEpoch struct {
	Idx int64
	Members map[string]float64						// 该轮次的所有成员初始权重
	Contribution map[string]*big.Float				// 该轮次的所有成员计算出得贡献度
}