package main

func main() {
	dnsMsg := dnsMessage{
		header: dnsHeader{
			id:      22,
			rd:      1,
			qdCount: 1,
		},
	}
	dnsMsg.encQuestionName("dns.google.com")
	_ = dnsMsg.packMessageQuerry()
}
