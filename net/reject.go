package net

type MsgReject struct {
	Message string
	Code    uint8
	Reason  string
	Data    [32]byte
}

func (m *MsgReject) Command() string {
	return NMT_REJECT
}

func (m *MsgReject) Read(h *NetHeader) {
	m.Message = h.ReadString()
	m.Code = h.ReadUint8()
	m.Reason = h.ReadString()
	h.ReadBytes(m.Data[:])
}

func (m *MsgReject) Write(h *NetHeader) {
	h.WriteString(m.Message)
	h.WriteUint8(m.Code)
	h.WriteString(m.Reason)
	h.WriteBytes(m.Data[:])
}

func NewMsgReject() *MsgReject {
	return &MsgReject{}
}
