/* SPDX-License-Identifier: MIT */

package device

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

const decoyServerBanner = "nginx/1.24.0"

// decoyServer listens on the same TCP port as the UDP WireGuard endpoint and
// responds as a plain nginx web server. This defeats active probing: TSPU
// connects over TCP to verify whether the port belongs to a VPN; a plausible
// HTTP response makes the port look like a legitimate web service.
type decoyServer struct {
	mu       sync.Mutex
	listener net.Listener
}

func (d *decoyServer) start(port uint16, logger *Logger) {
	d.mu.Lock()
	old := d.listener
	d.listener = nil
	d.mu.Unlock()

	if old != nil {
		old.Close()
	}

	if port == 0 {
		return
	}

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Verbosef("Decoy TCP server failed to bind :%d: %v", port, err)
		return
	}

	d.mu.Lock()
	d.listener = ln
	d.mu.Unlock()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go handleDecoyConn(conn)
		}
	}()
	logger.Verbosef("Decoy TCP server listening on :%d", port)
}

func (d *decoyServer) stop() {
	d.mu.Lock()
	ln := d.listener
	d.listener = nil
	d.mu.Unlock()

	if ln != nil {
		ln.Close()
	}
}

func handleDecoyConn(conn net.Conn) {
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(8 * time.Second))

	buf := make([]byte, 512)
	n, _ := conn.Read(buf)

	// brief pause to simulate nginx request processing
	time.Sleep(50 * time.Millisecond)

	var status, body string
	if n >= 4 && isHTTPRequest(buf[:n]) {
		status = "200 OK"
		body = "<!DOCTYPE html><html><head><title>Welcome to nginx!</title></head>" +
			"<body><center><h1>Welcome to nginx!</h1></center>" +
			"<hr><center>" + decoyServerBanner + "</center></body></html>"
	} else {
		status = "400 Bad Request"
		body = "<html><head><title>400 Bad Request</title></head>" +
			"<body><center><h1>400 Bad Request</h1></center>" +
			"<hr><center>" + decoyServerBanner + "</center></body></html>"
	}

	response := fmt.Sprintf(
		"HTTP/1.1 %s\r\nServer: %s\r\nDate: %s\r\nContent-Type: text/html\r\nContent-Length: %d\r\nConnection: close\r\n\r\n%s",
		status, decoyServerBanner,
		time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT"),
		len(body), body,
	)
	conn.Write([]byte(response))
}

func isHTTPRequest(buf []byte) bool {
	s := string(buf)
	return strings.HasPrefix(s, "GET ") || strings.HasPrefix(s, "POST ") ||
		strings.HasPrefix(s, "HEAD ") || strings.HasPrefix(s, "PUT ") ||
		strings.HasPrefix(s, "DELETE ") || strings.HasPrefix(s, "OPTIONS ")
}
