package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	RR_TYPE_HOST_ADDRESS = 1
	RR_CLASS_IN          = 1
	FLAGS_RD             = 1 << 8
)

type dnsHeader struct {
	id      uint16
	flags   uint16
	qdCount uint16
	anCount uint16
	nsCount uint16
	arCount uint16
}

type dnsQuestion struct {
	name  []uint8
	qType uint16
	class uint16
}

type resourceRecord struct {
	name     []uint8
	rrType   uint16
	class    uint16
	ttl      uint32
	rdLength uint16
	rData    []uint8
}

type dnsAnswer []resourceRecord

type dnsAuthority []resourceRecord

type dnsAdditional []resourceRecord

type dnsMessage struct {
	header     dnsHeader
	question   dnsQuestion
	answer     dnsAnswer
	authority  dnsAuthority
	additional dnsAdditional
}

func (m *dnsMessage) encQuestionName(host string) {
	var buff bytes.Buffer

	for _, part := range append(bytes.Split([]byte(host), []byte(".")), []byte{}) {
		buff.WriteByte(byte(len(part)))
		buff.Write(part)
	}
	m.question.name = buff.Bytes()
	//qType, class - hardcoded for now using constants
	m.question.qType = RR_TYPE_HOST_ADDRESS
	m.question.class = RR_CLASS_IN
}

func (m dnsMessage) packMessageQuerryBinary() []byte {
	var buff bytes.Buffer
	binary.Write(&buff, binary.BigEndian, m.header)
	binary.Write(&buff, binary.BigEndian, m.question.name[:])
	binary.Write(&buff, binary.BigEndian, m.question.qType)
	binary.Write(&buff, binary.BigEndian, m.question.class)

	fmt.Printf("%s\n", fmt.Sprintf("%0x", buff.Bytes()))
	return buff.Bytes()
}
