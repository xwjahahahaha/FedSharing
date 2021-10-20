package tx

import (
	"fmt"
	"strings"
)

type Input struct {
	OutputTxHash []byte
	OutputIdx int

	Signature []byte
	PublicKey []byte
}

// String
// @Description: 标准输出Input
// @receiver in
// @return string
func (in *Input) String() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("\t\t\t\tOutputTxHash: %x\n", in.OutputTxHash))
	lines = append(lines, fmt.Sprintf("\t\t\t\tOutputIdx: %d\n", in.OutputIdx))
	lines = append(lines, fmt.Sprintf("\t\t\t\tSignature: %x\n", in.Signature))
	lines = append(lines, fmt.Sprintf("\t\t\t\tPubKey: %x\n", in.PublicKey))
	return strings.Join(lines, "")
}