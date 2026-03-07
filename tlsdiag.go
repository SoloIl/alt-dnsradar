package main

import (
	"crypto/tls"
	"net"
	"time"
)

func tlsDiagnosticTimeout() time.Duration {
	return 3 * time.Second
}

func testTLSProbe(ip string, domain string) TLSStatus {
	timeout := tlsDiagnosticTimeout()

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
		if ne, ok := err.(net.Error); ok && ne.Timeout() {
			return TLSStatusTimeout
		}

		return TLSStatusFail
	}

	_ = conn.Close()
	return TLSStatusOK
}
