package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"time"
)

const HEADER_SIZE = 12

type Header struct {
	ID    uint16 // Packet Identifier (ID)
	Flags uint16
	// QR     Query/Response Indicator (QR)
	// OPCODE Operation Code (OPCODE)
	// AA     Authoritative Answer (AA)
	// TC     Truncation (TC)
	// RD     Recursion Desired (RD)
	// RA     Recursion Available (RA)
	// Z      Reserved (Z)
	// RCODE  Response Code (RCODE)
	QDCOUNT uint16 // Question Count (QDCOUNT)
	ANCOUNT uint16 // Answer Record Count (ANCOUNT)
	NSCOUNT uint16 // Authority Record Count (NSCOUNT)
	ARCOUNT uint16 // Additional Record Count (ARCOUNT)
}

type Question struct {
	QNAME  []byte
	QTYPE  uint16
	QCLASS uint16
}

type Answer struct {
	NAME     []byte
	TYPE     uint16
	CLASS    uint16
	TTL      uint32
	RDLENGTH uint16
	RDATA    []byte
}

type DNSQuery struct {
	Header
	Question
}

type DNSReply struct {
	Header
	Question
	Answer
}

func (h Header) Serialize() []byte {
	buf := make([]byte, HEADER_SIZE)
	binary.BigEndian.PutUint16(buf[0:2], h.ID)
	binary.BigEndian.PutUint16(buf[2:4], h.Flags)
	binary.BigEndian.PutUint16(buf[4:6], h.QDCOUNT)
	binary.BigEndian.PutUint16(buf[6:8], h.ANCOUNT)
	binary.BigEndian.PutUint16(buf[8:10], h.NSCOUNT)
	binary.BigEndian.PutUint16(buf[10:12], h.ARCOUNT)

	return buf
}

func NewHeader(buf []byte) (Header, error) {
	var h Header
	r := bytes.NewReader(buf)
	err := binary.Read(r, binary.BigEndian, &h)
	if err != nil {
		fmt.Printf("Failed to deserialize header: %s", err)
		return Header{}, err
	}
	return h, nil
}

func (q Question) Serialize() []byte {
	buf := make([]byte, len(q.QNAME)+4)
	copy(buf, q.QNAME)
	offset := len(q.QNAME)
	binary.BigEndian.PutUint16(buf[offset:], q.QTYPE)
	binary.BigEndian.PutUint16(buf[offset+2:], q.QCLASS)

	return buf
}

func NewQuestion(buf []byte) (Question, int, error) {
	nullIndex := bytes.IndexByte(buf, 0)
	if nullIndex == -1 {
		return Question{}, 0, errors.New("no null byte found in question buffer")
	}
	qnameData := buf[:nullIndex+1]
	offset := len(qnameData)
	questionSize := offset + 4
	if len(buf) < questionSize {
		return Question{}, 0, errors.New("question buffer is too small")
	}
	return Question{
		QNAME:  qnameData,
		QTYPE:  binary.BigEndian.Uint16(buf[offset : offset+2]),
		QCLASS: binary.BigEndian.Uint16(buf[offset+2 : questionSize]),
	}, questionSize, nil
}

func (a Answer) Serialize() []byte {
	a.RDLENGTH = uint16(len(a.RDATA))
	buf := make([]byte, len(a.NAME)+10+int(a.RDLENGTH))
	copy(buf, a.NAME)
	offset := len(a.NAME)
	binary.BigEndian.PutUint16(buf[offset:], a.TYPE)
	binary.BigEndian.PutUint16(buf[offset+2:], a.CLASS)
	binary.BigEndian.PutUint32(buf[offset+4:], a.TTL)
	binary.BigEndian.PutUint16(buf[offset+8:], a.RDLENGTH)
	copy(buf[len(a.NAME)+10:], a.RDATA)

	return buf
}

func NewAnswer(aname []byte, atype uint16, aclass uint16, attl uint32) Answer {
	// TODO: implement logic for what answer, IP address to return, etc.
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: 5 * time.Second}
			return d.DialContext(ctx, "udp", "8.8.8.8:53")
		},
	}
	domain := ParseDomainName(aname)
	ips, _ := r.LookupHost(context.Background(), domain)
	addr, _ := netip.ParseAddr(ips[0]) // TODO: it currently takes only the first item, was set up like that just for tests
	octets := addr.As4()

	return Answer{
		NAME:     aname,
		TYPE:     atype,
		CLASS:    aclass,
		TTL:      attl,
		RDLENGTH: 0,
		RDATA:    octets[:],
	}
}
