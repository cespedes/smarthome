package smarthome

import (
	"log"
	"net"
)

const coiotPort = 5683

type CoIoT struct {
	conn *net.UDPConn
	//mcast *net.UDPConn
}

const (
	TypeConfirmable     = 0
	TypeNonConfirmable  = 1
	TypeAcknowledgement = 2
	TypeReset           = 3
)

type CoIoToption struct {
	Number int
	Value  []byte
}

type CoIoTpacket struct {
	RemoteAddr *net.UDPAddr
	Version    int
	Type       int
	Code       [2]int
	MessageID  uint16
	Token      []byte
	Options    []CoIoToption
	Payload    []byte
	Path       []string
}

func CoIoTinit() (*CoIoT, error) {
	var c CoIoT
	var err error

	log.Printf("CoIoT: listening on UDP port %d.", coiotPort)
	c.conn, err = net.ListenUDP("udp", &net.UDPAddr{Port: coiotPort})
	if err != nil {
		return nil, err
	}

	//log.Printf("CoIoT: listening on UDP multicast addr 224.0.1.187:%d.", coiotPort)
	//c.mcast, err = net.ListenMulticastUDP("udp", nil, &net.UDPAddr{IP: net.IPv4(224, 0, 1, 187), Port: coiotPort})
	//if err != nil {
	//	return nil, err
	//}
	return &c, nil
}

func (c *CoIoT) Read() *CoIoTpacket {
	data := make([]byte, 65*1024)
	n, remoteAddr, err := c.conn.ReadFromUDP(data)
	if err != nil {
		log.Fatal(err)
	}
	data = data[:n]

	if n < 4 { // short packet
		return c.Read()
	}

	var p CoIoTpacket
	p.RemoteAddr = remoteAddr
	p.Version = int(data[0] >> 6)
	p.Type = int((data[0] >> 4) & 0x3)
	tokenLength := int(data[0] & 0xF)
	p.Code[0] = int(data[1] >> 5)
	p.Code[1] = int(data[1] & 0x1F)
	p.MessageID = uint16(data[2])<<8 | uint16(data[3])
	n = 4 + tokenLength
	p.Token = data[4:n]
	var lastNumber = 0
	for {
		if n >= len(data) {
			break
		}
		if data[n] == 0xFF {
			n++
			break
		}
		var o CoIoToption
		var delta = int(data[n] >> 4)
		var length = int(data[n] & 0xF)
		n++
		if delta == 13 {
			delta = 13 + int(data[n])
			n++
		} else if delta == 14 {
			delta = 269 + int(data[n])<<8 + int(data[n+1])
			n += 2
		}
		if length == 13 {
			length = 13 + int(data[n])
			n++
		} else if length == 14 {
			length = 269 + int(data[n])<<8 + int(data[n+1])
			n += 2
		}
		lastNumber += delta
		o.Number = lastNumber
		o.Value = data[n : n+length]
		//log.Printf("Found option: delta=%d number=%d length=%d value=%q", delta, o.Number, length, string(o.Value))
		p.Options = append(p.Options, o)
		n += length
	}
	if n < len(data) {
		p.Payload = data[n:]
	}

	for _, o := range p.Options {
		if o.Number < 11 {
			continue
		}
		if o.Number > 11 {
			break
		}
		p.Path = append(p.Path, string(o.Value))
	}

	/*
	if p.Code[0] != 0 || p.Code[1] != 30 { // unsupported CoIoT code
		return &p
	}
	if p.Version != 2 { // unsupported CoIoT version
		return &p
	}
	*/
	return &p
}
