package core

//

type MsgFeeFilter struct {
	FeeRate uint64
}

func (m *MsgFeeFilter) Command() string {
	return NMT_FEEFILTER
}

func (m *MsgFeeFilter) Read(h *NetHeader) {
	m.FeeRate = h.ReadUInt64()
}

func (m *MsgFeeFilter) Write(h *NetHeader) {
	h.WriteUInt64(m.FeeRate)
}

func NewMsgFeeFilter() *MsgFeeFilter {
	return &MsgFeeFilter{}
}

//

type MsgSendHeaders struct {
}

func (m *MsgSendHeaders) Command() string {
	return NMT_SENDHEADERS
}

func (m *MsgSendHeaders) Read(h *NetHeader) {
	//no payload
}

func (m *MsgSendHeaders) Write(h *NetHeader) {
	//no payload
}

func NewMsgSendHeaders() *MsgSendHeaders {
	return &MsgSendHeaders{}
}

//

type MsgAddr struct {
	Num   uint64
	Addrs []Address
}

func (m *MsgAddr) Command() string {
	return NMT_ADDR
}

func (m *MsgAddr) Read(h *NetHeader) {
	siz := GetAddressSize()
	num, l := h.ReadVarInt()
	m.Num = num
	size := (h.Len() - uint32(l)) / uint32(siz)
	m.Addrs = make([]Address, size)
	for i, _ := range m.Addrs {
		m.Addrs[i].Read(h, true)
	}
}

func (m *MsgAddr) Write(h *NetHeader) {
	//no payload
}

func NewMsgAddr() *MsgAddr {
	return &MsgAddr{}
}

//
type MsgGetAddr struct {
}

func (m *MsgGetAddr) Command() string {
	return NMT_GETADDR
}

func (m *MsgGetAddr) Read(h *NetHeader) {
	//no payload
}

func (m *MsgGetAddr) Write(h *NetHeader) {
	//no payload
}

func NewMsgGetAddr() *MsgGetAddr {
	return &MsgGetAddr{}
}
