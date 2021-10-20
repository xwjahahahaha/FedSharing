package node


type CollectClientsMsg struct {
	TaskID int
	NowClientCount int
	NowClients []string
	AddrFrom string
}

type GlobalEpochMsg struct {
	GlobalEpoch int
	MultiAddr string
	AddrFrom string
}

type EstablishStreamMsg struct {
	MultiAddr string
	ClientID int
	AddrFrom string
}

type ClientDiffMsg struct {
	GlobalEpoch int
	ClientID int
	DiffBlock []byte
	AddrFrom string
}