package bw2bind

import (
	"bufio"
	"errors"
	"net"
	"time"

	log "github.com/cihub/seelog"
	"github.com/immesys/bw2/objects"
)

type BW2Client struct {
	c         net.Conn
	out       *bufio.Writer
	in        *bufio.Reader
	remotever string
}

func (cl *BW2Client) Publish()
func Connect(to string) (*BW2Client, error) {
	conn, err := net.Dial("tcp", to)
	if err != nil {
		return nil, err
	}
	rv := &BW2Client{c: conn,
		out: bufio.NewWriter(conn),
		in:  bufio.NewReader(conn)}

	//As a bit of a sanity check, we read the first frame, which is the
	//server HELO message
	ok := make(chan bool, 1)
	go func() {
		helo, err := objects.LoadFrameFromStream(rv.in)
		if err != nil {
			log.Error("Malformed HELO frame: ", err)
			ok <- false
			return
		}
		if helo.Cmd != objects.CmdHello {
			log.Error("Frame not HELO")
			ok <- false
			return
		}
		rver, hok := helo.GetFirstHeader("version")
		if !hok {
			log.Error("Frame has no version")
			ok <- false
			return
		}
		rv.remotever = rver
		log.Info("Connected to BOSSWAVE router version ", rver)
		ok <- true
	}()
	select {
	case okv := <-ok:
		if okv {
			return rv, nil
		}
		return nil, errors.New("Bad router")
	case _ = <-time.After(5 * time.Second):
		log.Error("Timeout on router HELO")
		conn.Close()
		return nil, errors.New("Timeout on HELO")
	}
}
