package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	dnsMsg := dnsMessage{
		header: DNSHeader{
			ID: 22,
			//set recursion desired bit
			Flags:   FLAGS_RD,
			QDCount: 1,
		},
	}
	dnsMsg.encQuestionName("dns.google.com")
	dnsQuery := dnsMsg.pack()

	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		fmt.Println("error while connecting to Google’s DNS: ", err)
		os.Exit(1)
	}

	_, err = conn.Write([]byte(dnsQuery))
	if err != nil {
		fmt.Println("error sending DNS query through socket: ", err)
		os.Exit(1)
	}

	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("failed to get response:", err)
		return
	}

	respMsg := dnsMessage{}
	respMsg.unpack(buf[:n])
	fmt.Println(respMsg.header.ID, respMsg.header.Flags, respMsg.header.QDCount, respMsg.header.ANCount, respMsg.header.NSCount, respMsg.header.ARCount)
	fmt.Println(respMsg.question)
	fmt.Println(respMsg.answer)
}
