package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

func testTLSProbe(ip string, domain string) {

	fmt.Print("Testing TLS handshake ")

	spin := []rune{'|', '/', '-', '\\'}

	for i := 0; i < 10; i++ {

		fmt.Printf("\rTesting TLS handshake %c", spin[i%4])

		time.Sleep(120 * time.Millisecond)
	}

	timeout := time.Duration(*flagSettings.RequestsTimeout) * time.Second

	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: timeout},
		"tcp",
		ip+":443",
		&tls.Config{
			ServerName:         domain,
			InsecureSkipVerify: true,
		},
	)

	if err != nil {

		fmt.Println("\rTLS interference likely")
		return
	}

	conn.Close()

	fmt.Println("\rTLS handshake OK")
}
