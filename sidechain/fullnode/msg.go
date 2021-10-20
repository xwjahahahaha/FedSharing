package fullnode

// VersionMsg 区块链版本信息，确定最长链
type VersionMsg struct {
	SideChainVersion int
	BestHeight int
	AddrFrom string
}

// GetBlockMsg 请求获取区块信息
type GetBlockMsg struct {
	HeightRange []int		// 需要的区块范围
	AddrFrom string
}

// BlockMsg 区块消息
type BlockMsg struct {
	Block []byte
	AddrFrom string
}