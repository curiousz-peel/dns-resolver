package main

import (
	"fmt"
	"strings"
)

const (
	RR_TYPE_HOST_ADDRESS = 1
	RR_CLASS_IN          = 1
)

type dnsHeader struct {
	id      uint16
	qr      uint8
	opCode  uint8
	aa      uint8
	tc      uint8
	rd      uint8
	ra      uint8
	z       uint8
	rCode   uint8
	qdCount uint16
	anCount uint16
	nsCount uint16
	arCount uint16
}

type dnsQuestion struct {
	name  []byte
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
	var stringEnc string

	for _, s := range strings.Split(host, ".") {
		stringEnc += fmt.Sprintf("%d%s", len(s), s)
	}
	stringEnc += "0"

	fmt.Println(stringEnc)
	m.question.name = []byte(stringEnc)
	//qType, class - TBD, hardcoded for now using constants
}

func byteSliceToByteString(bSlice []byte) string {
	var bString string
	for _, b := range bSlice {
		bString += fmt.Sprintf("%0b", b)
	}
	return bString
}

func (m dnsMessage) packMessageQuerry() string {
	msgQuerry := strings.Join([]string{
		//2 bytes for the ID
		fmt.Sprintf("%016b", m.header.id),
		//2 bytes for flags
		fmt.Sprintf("%01b", m.header.qr),
		fmt.Sprintf("%04b", m.header.opCode),
		fmt.Sprintf("%01b", m.header.aa),
		fmt.Sprintf("%01b", m.header.tc),
		fmt.Sprintf("%01b", m.header.rd),
		fmt.Sprintf("%01b", m.header.ra),
		fmt.Sprintf("%03b", m.header.z),
		fmt.Sprintf("%04b", m.header.rCode),
		//2 bytes for no. of questions,
		//no. of resource records in the answer, authority and additional sections each
		fmt.Sprintf("%016b", m.header.qdCount),
		fmt.Sprintf("%016b", m.header.anCount),
		fmt.Sprintf("%016b", m.header.arCount),
		//encoded host name bits
		byteSliceToByteString(m.question.name[:]),
		//2 bytes for the query type
		fmt.Sprintf("%016b", RR_TYPE_HOST_ADDRESS),
		//2 bytes for the query class
		fmt.Sprintf("%016b", RR_CLASS_IN),
	},
		"")

	return msgQuerry
}
