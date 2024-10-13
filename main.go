package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	dnsMsg := dnsMessage{
		header: dnsHeader{
			id: 22,
			//set recursion desired bit
			flags:   FLAGS_RD,
			qdCount: 1,
		},
	}
	dnsMsg.encQuestionName("dns.google.com")
	dnsQuerry := dnsMsg.packMessageQuerryBinary()

	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		fmt.Println("error while connecting to Google’s DNS: ", err)
		os.Exit(1)
	}

	_, err = conn.Write([]byte(dnsQuerry))
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

	fmt.Printf("%0x\n", buf[:n])
}
