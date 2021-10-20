package pofs

import (
	"crypto/sha256"
	"fedSharing/sidechain/task"
	"log"
	"math/big"
	"math/rand"
	"strings"
)

const PREC uint = 100				// 精度

func CalculateSV(task *task.Task, epochIdx int64) map[string]*big.Float {
	var memberNames []string
	for m := range task.Epochs[epochIdx].Members {
		memberNames = append(memberNames, m)
	}
	for member := range task.Epochs[epochIdx].Members {
		// 1. 计算S_i：包含该成员的所有集合
		s := getAllSubsetsAboutOne(memberNames, member)
		// 2. 对于每个集合：计算其v(s) 和 v(s\m)
		for _, set := range s {
			var contribution big.Float
			contribution.Sub(VSet(set), VSet(RemoveItem(set, member)))
			if contribution.Cmp(big.NewFloat(0)) == -1 {
				contribution.SetFloat64(0)
			}
			// 3. 计算w(|s|)
			weight := W(len(memberNames), len(set))
			// 4. 设置结果
			newVarphi := weight.Mul(weight, &contribution)
			if _, has := task.Epochs[epochIdx].Contribution[member]; !has {
				task.Epochs[epochIdx].Contribution[member] = newVarphi
			}else {
				task.Epochs[epochIdx].Contribution[member].Add(task.Epochs[epochIdx].Contribution[member], newVarphi)
			}
		}
	}
	return task.Epochs[epochIdx].Contribution
}

// makeAllSubsets
// @Description: 获取当前集合的所有子集
// @param collection	集合
// @param hasNull	结果中是否包含空集
// @return set
func makeAllSubsets(collection []string, hasNull bool) (set [][]string) {
	// 采用位运算实现整形数与集合的映射
	n := len(collection)
	max := (1 << n) -1
	var binNum int
	if !hasNull {
		binNum = 1
	}
	for ; binNum <= max; binNum++ {
		// 根据当前数的二进制判断取元素
		var subset []string
		for i := 0; i < n; i ++ {			// 遍历集合所有元素下标
			if (1 << i) & binNum >= 1 {		// 注意：这里判断要写>=1或者!=0
				subset = append(subset, collection[i])
			}
		}
		set = append(set, subset)
	}
	return
}

// getAllSubsetsAboutOne
// @Description: 获取集合的所有子集中含当前元素的子集
// @param collection 集合
// @param m	元素
// @return set 所有满足要求的子集
func getAllSubsetsAboutOne(collection []string, m string) (set [][]string) {
	if !IsItem(collection, m) {
		log.Panic("The collection not have this item")
	}
	// 去掉自身
	collection = RemoveItem(collection, m)
	// 求所有子集
	subsets := makeAllSubsets(collection, true)
	for _, subset := range subsets {
		subset = append(subset, m)
		set = append(set, subset)
	}
	return
}

// VSet
// @Description: 计算特征函数
// @param set 数据集
// @return *big.Float 联盟贡献总量
func VSet(set []string) *big.Float {
	// 计算一个集合的总体贡献度
	// 计算对模型的损失  //TODO 访问Pool Manager
	// 先用随机数模拟 保证联合比单独更好
	n := len(set)
	res := big.NewFloat(0).SetPrec(PREC)
	if n == 0 {
		return res.SetFloat64(0)
	}
	if n == 1 {
		return res.SetFloat64(float64(sha256.Sum256([]byte(set[0]))[0]))
	}
	var sum big.Float
	for _, item := range set {
		sum.Add(&sum, big.NewFloat(float64(sha256.Sum256([]byte(item))[0])))
	}
	rand.Seed(1)
	return res.SetFloat64(float64(rand.Intn(30))).Add(res, &sum)
}

// W
// @Description: 计算权重
// @param n 总大小
// @param s 子联盟大小
// @return *big.Float 权重
func W(n, s int) *big.Float {
	molecular, denominator := big.NewFloat(1).SetPrec(PREC), big.NewFloat(1).SetPrec(PREC)
	for i := 1; i <= s-1; i++ {
		molecular.Mul(molecular, big.NewFloat(float64(i)))
	}
	for i := 1; i <= n-s; i++ {
		molecular.Mul(molecular, big.NewFloat(float64(i)))
	}
	for i := 1; i <= n; i++ {
		denominator.Mul(denominator, big.NewFloat(float64(i)))
	}
	return molecular.Quo(molecular, denominator)
}

// IsItem
// @Description: 判断是否是切片中的元素
// @param c
// @param s
// @return bool
func IsItem(c []string, s string) bool {
	for _, str := range c {
		if strings.Compare(str, s) == 0 {
			return true
		}
	}
	return false
}

// RemoveItem
// @Description: 删除掉string数组中的一个元素
// @param c	string数组
// @param s 元素
// @return []string
func RemoveItem(c []string, s string) []string {
	n := len(c)
	for i, m := range c {
		if strings.Compare(s, m) == 0 {
			c[i], c[n-1] = c[n-1], c[i]
			c = c[:n-1]
			break
		}
	}
	return c
}
