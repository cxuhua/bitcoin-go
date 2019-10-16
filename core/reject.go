package core

type MsgReject struct {
	Message string
	Code    uint8
	Reason  string
}

func (m *MsgReject) Command() string {
	return NMT_REJECT
}

func (m *MsgReject) Read(h *NetHeader) {
	m.Message = h.ReadString()
	m.Code = h.ReadUint8()
	m.Reason = h.ReadString()
}

func (m *MsgReject) Write(h *NetHeader) {
	h.WriteString(m.Message)
	h.WriteUint8(m.Code)
	h.WriteString(m.Reason)
}

func NewMsgReject() *MsgReject {
	return &MsgReject{}
}
