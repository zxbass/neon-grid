package main

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
)

// wsAccept computes the RFC 6455 Sec-WebSocket-Accept value.
func wsAccept(key string) string {
	h := sha1.Sum([]byte(key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	return base64.StdEncoding.EncodeToString(h[:])
}

// wsTextFrame builds an unmasked server->client text frame.
func wsTextFrame(msg string) []byte {
	b := []byte{0x81, byte(len(msg))}
	return append(b, msg...)
}

// ---------------------------------------------------------------- 041 ping

func (s *set) gen041() {
	id := "041"
	// file-based port scan: hosts.txt lists the targets, banners.txt holds
	// the banner for each OPEN port (ports missing from it are CLOSED).
	hosts := "127.0.0.1 23\n127.0.0.1 9999\n127.0.0.1 8080\n127.0.0.1 1200\n"
	s.text(id, "hosts.txt", hosts)

	banners := "23 OMEGA-DYNE GATEWAY v2.1 // UNAUTHORIZED ACCESS IS FELONY\n" +
		"8080 Apache/2.4.41 (Unix) mod_ssl/2.4.41\n" +
		"1200 NEON-CHAT/0.9 READY\n"
	s.text(id, "banners.txt", banners)

	open := map[string]string{}
	for _, l := range strings.Split(strings.TrimSpace(banners), "\n") {
		port := strings.Fields(l)[0]
		open[port] = strings.Join(strings.Fields(l)[1:], " ")
	}
	var p1, p2 []string
	for _, l := range strings.Split(strings.TrimSpace(hosts), "\n") {
		f := strings.Fields(l)
		host, port := f[0], f[1]
		if b, ok := open[port]; ok {
			p1 = append(p1, fmt.Sprintf("%s %s    -> OPEN", host, port))
			p2 = append(p2, fmt.Sprintf("BANNER: %s", b))
		} else {
			p1 = append(p1, fmt.Sprintf("%s %s  -> CLOSED", host, port))
		}
	}
	s.expected(id, strings.Join(p1, "\n"), strings.Join(p2, "\n"))
}

// ---------------------------------------------------------------- 042 echo

func (s *set) gen042() {
	id := "042"
	log := "12:01:22 127.0.0.1:54321 -> \"HELLO crow\"\n" +
		"12:01:22 127.0.0.1:54321 -> \"BYE\"\n" +
		"12:01:25 127.0.0.1:54322 -> \"HELLO zen\"\n" +
		"12:01:25 127.0.0.1:54322 -> \"ping\"\n" +
		"12:01:26 127.0.0.1:54322 -> \"BYE\"\n"
	s.text(id, "echo.log", log)

	var lines []string
	for _, l := range strings.Split(strings.TrimSpace(log), "\n") {
		lines = append(lines, l)
	}
	p1 := fmt.Sprintf("CONNECTIONS: 2\nEXCHANGES: 5\nLAST: %s", lines[len(lines)-1])
	p2 := strings.Join(lines, "\n")
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 043 http

func (s *set) gen043() {
	id := "043"
	resp := "HTTP/1.1 200 OK\r\n" +
		"server: omegadyn/2.1\r\n" +
		"content-type: text/plain\r\n" +
		"content-length: 47\r\n" +
		"\r\n" +
		"MESSAGE FROM THE GRID: WE ARE MANY."
	s.text(id, "response.txt", resp)

	head, body, _ := strings.Cut(resp, "\r\n\r\n")
	status := strings.Fields(head)[1] + " " + strings.Fields(head)[2]
	var hdrs []string
	for _, l := range strings.Split(head, "\r\n")[1:] {
		hdrs = append(hdrs, strings.ToLower(l))
	}
	p1 := fmt.Sprintf("STATUS: %s\nHEADERS:\n%s\nBODY:\n%s", status, strings.Join(hdrs, "\n"), body)
	p2 := "FLAG: FLAG{HTTP_IS_OLD_SCHOOL}"
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 044 dns

func (s *set) gen044() {
	id := "044"
	qname := func(name string) []byte {
		var b []byte
		for _, lbl := range strings.Split(name, ".") {
			b = append(b, byte(len(lbl)))
			b = append(b, lbl...)
		}
		return append(b, 0)
	}
	var pkt []byte
	pkt = append(pkt, 0x12, 0x34, 0x81, 0x80, 0x00, 0x01, 0x00, 0x03, 0x00, 0x00, 0x00, 0x00)
	q := qname("example.com")
	pkt = append(pkt, q...)
	pkt = append(pkt, 0x00, 0x01, 0x00, 0x01)

	namePtr := func(off int) []byte { return []byte{0xC0, byte(off)} }
	rrHeader := func(name []byte, typ uint16) []byte {
		h := append([]byte{}, name...)
		return append(h, byte(typ>>8), byte(typ), 0x00, 0x01, 0x00, 0x00, 0x00, 0x3C)
	}

	// answer 1: example.com CNAME edge.example.net
	rdata := append([]byte{0x04}, []byte("edge")...)
	rdata = append(rdata, 0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 0x03, 'n', 'e', 't', 0x00)
	a1 := rrHeader(namePtr(12), 5)
	a1 = append(a1, byte(len(rdata)>>8), byte(len(rdata)))
	a1 = append(a1, rdata...)
	rdataOff := len(pkt) + len(a1) - len(rdata)
	pkt = append(pkt, a1...)

	// answer 2: edge.example.net A 192.0.2.1 (NAME -> ptr to "edge" label in a1)
	a2 := rrHeader(namePtr(rdataOff), 1)
	a2 = append(a2, 0x00, 0x04, 192, 0, 2, 1)
	pkt = append(pkt, a2...)

	// answer 3: example.com A 93.184.216.34
	a3 := rrHeader(namePtr(12), 1)
	a3 = append(a3, 0x00, 0x04, 93, 184, 216, 34)
	pkt = append(pkt, a3...)
	s.file(id, "reply.bin", pkt)

	p1 := "ID=0x1234 RCODE=0\nedge.example.net. -> 192.0.2.1\nexample.com. -> 93.184.216.34"
	p2 := "example.com. CNAME -> edge.example.net.\nedge.example.net. A -> 192.0.2.1\nexample.com. A -> 93.184.216.34"
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 045 chat

func (s *set) gen045() {
	id := "045"
	transcript := "[CROW] ready\n" +
		"[ZEN] ready\n" +
		"[CROW] target locked\n" +
		"[ZEN] copy that\n" +
		"[system] CROW lost link\n"
	s.text(id, "transcript.txt", transcript)

	var msgs []string
	var sys []string
	for _, l := range strings.Split(strings.TrimSpace(transcript), "\n") {
		if strings.HasPrefix(l, "[system]") {
			sys = append(sys, l)
		} else {
			msgs = append(msgs, l)
		}
	}
	p1 := fmt.Sprintf("MESSAGES: %d\nLAST: %s", len(msgs), msgs[len(msgs)-1])
	p2 := strings.Join(sys, "\n")
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 046 ftp

func (s *set) gen046() {
	id := "046"
	sess := "USER crow -> 200 OK\n" +
		"LIST\n" +
		"boot.bin\t512\n" +
		"map.raw\t4096\n" +
		"200 END\n" +
		"GET boot.bin -> 512 bytes\n" +
		"USER agent.ops -> 200 OK\n" +
		"GET vault.key -> 2048 bytes\n"
	s.text(id, "session.txt", sess)

	lines := strings.Split(strings.TrimSpace(sess), "\n")
	var p1, p2 []string
	for _, l := range lines {
		if strings.HasPrefix(l, "USER") {
			if strings.Contains(l, "agent.ops") {
				p2 = append(p2, l)
			} else {
				p1 = append(p1, l)
			}
		} else if strings.HasPrefix(l, "GET") {
			p1 = append(p1, l)
		}
	}
	s.expected(id, strings.Join(p1, "\n"), strings.Join(p2, "\n"))
}

// ---------------------------------------------------------------- 047 websocket

func (s *set) gen047() {
	id := "047"
	key := "dGhlIHNhbXBsZSBub25jZQ=="
	accept := wsAccept(key)
	s.text(id, "handshake.txt", "key: "+key+"\n")
	frame := wsTextFrame("MERCURY_IS_AWAKE")
	s.file(id, "frame.bin", frame)

	p1 := "ACCEPT MATCH: " + accept
	p2 := "FRAME: MERCURY_IS_AWAKE"
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 048 scanner

func (s *set) gen048() {
	id := "048"
	// deterministic scan result: open ports listed in ports.txt, part1
	// reports them, part2 reports count + a fake elapsed time.
	ports := "22\n80\n1200\n8080\n"
	s.text(id, "ports.txt", ports)

	var open []string
	for _, l := range strings.Split(strings.TrimSpace(ports), "\n") {
		open = append(open, l)
	}
	p1 := "OPEN: " + strings.Join(open, " ")
	p2 := fmt.Sprintf("SCANNED 1024 ports in 213ms\nOPEN: %s", strings.Join(open, " "))
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 049 proxy

func (s *set) gen049() {
	id := "049"
	// traffic.log: bytes logged by the proxy in both directions
	log := "< 6f 6d 65 67 61 0a  |omega.|\n" +
		"> 6f 6b 0a            |ok.|\n" +
		"< 6f 6d 65 67 61 0a  |omega.|\n" +
		"> 6f 6b 0a            |ok.|\n"
	s.text(id, "traffic.log", log)

	lines := strings.Split(strings.TrimSpace(log), "\n")
	down := 0
	for _, l := range lines {
		if strings.HasPrefix(l, "<") {
			down++
		}
	}
	p1 := fmt.Sprintf("LOGGED: %d lines (%d client->server)", len(lines), down)
	p2 := "ROUNDTRIP OK"
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 050 banners

func (s *set) gen050() {
	id := "050"
	banners := "127.0.0.1 23   OMEGA-DYNE GATEWAY v2.1\n" +
		"127.0.0.1 8080  Apache/2.4.41 (Unix) mod_ssl/2.4.41\n" +
		"127.0.0.1 1200  NEON-CHAT/0.9 READY\n" +
		"127.0.0.1 2222  OpenSSH_9.6 Ubuntu-3\n"
	s.text(id, "banners.txt", banners)

	names := map[string]string{
		"omega":    "OMEGA-DYNE",
		"apache":   "APACHE_HTTPD",
		"neon-chat": "NEON_CHAT",
		"openssh":  "OPENSSH",
	}
	counts := map[string]int{}
	var p1 []string
	for _, l := range strings.Split(strings.TrimSpace(banners), "\n") {
		f := strings.Fields(l)
		host, port := f[0], f[1]
		text := strings.Join(f[2:], " ")
		ven := "UNKNOWN"
		lt := strings.ToLower(text)
		for k, v := range names {
			if strings.Contains(lt, k) {
				ven = v
				break
			}
		}
		counts[ven]++
		p1 = append(p1, fmt.Sprintf("%s:%s   %s", host, port, text))
	}
	var vend []string
	for v := range counts {
		vend = append(vend, v)
	}
	sort.Strings(vend)
	var p2 []string
	for _, v := range vend {
		p2 = append(p2, fmt.Sprintf("%s: %d", v, counts[v]))
	}
	s.expected(id, strings.Join(p1, "\n"), strings.Join(p2, "\n"))
}

func init() {
	register("041", (*set).gen041)
	register("042", (*set).gen042)
	register("043", (*set).gen043)
	register("044", (*set).gen044)
	register("045", (*set).gen045)
	register("046", (*set).gen046)
	register("047", (*set).gen047)
	register("048", (*set).gen048)
	register("049", (*set).gen049)
	register("050", (*set).gen050)
}