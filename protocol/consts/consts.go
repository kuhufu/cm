package consts

const (
	CmdUnknown    = 0
	CmdAuth       = 1
	CmdPush       = 2
	CmdHeartbeat  = 3
	CmdClose      = 4
	CmdServerPush = 5
)

const (
	KB = 1 << 10
	MB = KB << 10
)

const (
	DefaultMagicNumber = 0x08
	MaxBodyLen         = 2 * MB
)
