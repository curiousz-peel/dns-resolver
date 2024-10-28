package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	RR_TYPE_HOST_ADDRESS = 1
	RR_CLASS_IN          = 1

	FLAGS_RD = 1 << 8
	FLAGS_QR = 1 << 15

	IS_COMPRESSED    = 1<<7 | 1<<6
	POINTER_VAL_MASK = 0xFFFF >> 2
)

type DNSHeader struct {
	ID      uint16
	Flags   uint16
	QDCount uint16
	ANCount uint16
	NSCount uint16
	ARCount uint16
}

type DNSQuestion struct {
	Name  []uint8
	QType uint16
	Class uint16
}

type ResourceRecord struct {
	Name     []uint8
	RRType   uint16
	Class    uint16
	TTL      uint32
	RDLength uint16
	RData    []uint8
}

type DNSAnswer []ResourceRecord

type DNSAuthority []ResourceRecord

type DNSAdditional []ResourceRecord

type dnsMessage struct {
	header     DNSHeader
	question   DNSQuestion
	answer     DNSAnswer
	authority  DNSAuthority
	additional DNSAdditional
}

func (m *dnsMessage) encQuestionName(host string) {
	var buff bytes.Buffer
	for _, part := range append(bytes.Split([]byte(host), []byte(".")), []byte{}) {
		buff.WriteByte(byte(len(part)))
		buff.Write(part)
	}
	m.question.Name = buff.Bytes()
	//qType, class - hardcoded for now using constants
	m.question.QType = RR_TYPE_HOST_ADDRESS
	m.question.Class = RR_CLASS_IN
}

func (m dnsMessage) pack() []byte {
	var buff bytes.Buffer
	binary.Write(&buff, binary.BigEndian, m.header)
	binary.Write(&buff, binary.BigEndian, m.question.Name[:])
	binary.Write(&buff, binary.BigEndian, m.question.QType)
	binary.Write(&buff, binary.BigEndian, m.question.Class)
	return buff.Bytes()
}

func (m *dnsMessage) unpack(resp []byte) error {
	var off int
	//unpack header
	off = unpackHeader(resp, m)

	//check if unpacked message is a response
	if m.header.Flags&FLAGS_QR == 0 {
		m = &dnsMessage{}
		return fmt.Errorf("QR bit not set, message is not a response")
	}

	//unpack question
	off = unpackQuestion(resp, off, m)

	//unpack answer
	var answer DNSAnswer
	for i := 0; i < int(m.header.ANCount); i++ {
		var rr ResourceRecord
		rr, off = unpackResourceName(resp, off)
		answer = append(answer, rr)
	}
	m.answer = answer

	//unpack authority
	var authority DNSAuthority
	for i := 0; i < int(m.header.NSCount); i++ {
		var rr ResourceRecord
		rr, off = unpackResourceName(resp, off)
		answer = append(answer, rr)
	}
	m.authority = authority

	//unpack additional
	var additional DNSAdditional
	for i := 0; i < int(m.header.ARCount); i++ {
		var rr ResourceRecord
		rr, off = unpackResourceName(resp, off)
		answer = append(answer, rr)
	}
	m.additional = additional
	return nil
}

func unpackName(resp []byte, off int) (name []byte, newOff int) {
	var buff bytes.Buffer
	var hadPointer bool
	var actualOff int
Outer:
	for {
		length := int(resp[off])
		switch {
		//if byte's 0 -> we reached the domain name end
		case length == 0:
			off++
			if !hadPointer {
				actualOff = off
			}
			break Outer
		//if first 2 bits are set -> compression is on, follow the pointer
		case length&IS_COMPRESSED == IS_COMPRESSED:
			if !hadPointer {
				hadPointer = true
				actualOff = off + 2
			}
			off = int(binary.BigEndian.Uint16(resp[off:]) & POINTER_VAL_MASK)
			continue
		//else read length no. bytes from the domain name and write to buffer
		default:
			off++
			buff.Write(resp[off : off+length])
			buff.WriteByte('.')
			off += length
		}
	}
	name = buff.Bytes()

	return name[:len(name)-1], actualOff
}

func unpackHeader(resp []byte, msg *dnsMessage) int {
	msg.header.ID = binary.BigEndian.Uint16(resp[0:])
	msg.header.Flags = binary.BigEndian.Uint16(resp[2:])
	msg.header.QDCount = binary.BigEndian.Uint16(resp[4:])
	msg.header.ANCount = binary.BigEndian.Uint16(resp[6:])
	msg.header.NSCount = binary.BigEndian.Uint16(resp[8:])
	msg.header.ARCount = binary.BigEndian.Uint16(resp[10:])

	//offset after parsing the header
	return 12
}

func unpackQuestion(resp []byte, currOff int, msg *dnsMessage) int {
	msg.question.Name, currOff = unpackName(resp, currOff)

	fmt.Printf("%s\n", fmt.Sprintf("%s\t%d", msg.question.Name, currOff))

	msg.question.QType = binary.BigEndian.Uint16(resp[currOff:])
	msg.question.Class = binary.BigEndian.Uint16(resp[currOff+2:])
	//increment offset to move past QType and QClass
	return currOff + 4
}

func unpackResourceName(resp []byte, off int) (ResourceRecord, int) {
	var rr ResourceRecord
	rr.Name, off = unpackName(resp, off)
	rr.RRType = binary.BigEndian.Uint16(resp[off:])
	rr.Class = binary.BigEndian.Uint16(resp[off+2:])
	rr.TTL = binary.BigEndian.Uint32(resp[off+4:])
	rr.RDLength = binary.BigEndian.Uint16(resp[off+8:])
	rr.RData = resp[off+10 : off+10+int(rr.RDLength)]
	off += 10 + int(rr.RDLength)
	return rr, off
}
