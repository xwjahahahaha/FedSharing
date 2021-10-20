package pofs

import (
	"fedSharing/sidechain/task"
	"math/big"
)

type PoFS struct {
	Difficulty float64
	Task task.Task
	EpochIdx int64
}

// NewPoFS
// @Description: 创建新的PoFS共识
// @param task	FL任务
// @param epochIdx	全局迭代轮次
// @return *PoFS
func NewPoFS(task task.Task, epochIdx int64) *PoFS {
	return &PoFS{
		Difficulty:  AssessDifficulty(),
		Task:        task,
		EpochIdx:    epochIdx,
	}
}

// Run
// @Description: 运算PoFS
// @receiver pofs
func (pofs *PoFS) Run() map[string]*big.Float {
	return CalculateSV(&pofs.Task, pofs.EpochIdx)
}


//TODO 评估调整当前难度
func AssessDifficulty() float64 {
	return 0.5
}

