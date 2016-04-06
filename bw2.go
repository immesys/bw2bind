package bw2bind

import (
	"bufio"
	"net"
	"sync"
)

type BW2Client struct {
	c            net.Conn
	out          *bufio.Writer
	in           *bufio.Reader
	remotever    string
	seqnos       map[int]chan *Frame
	olock        sync.Mutex
	curseqno     uint32
	defAutoChain *bool
}

//Sends a request frame and returns a  chan that contains all the responses.
//Automatically closes the returned channel when there are no more responses.
func (cl *BW2Client) transact(req *Frame) chan *Frame {
	seqno := req.SeqNo
	inchan := make(chan *Frame, 3)
	outchan := make(chan *Frame, 3)
	cl.olock.Lock()
	cl.seqnos[seqno] = inchan
	req.WriteToStream(cl.out)
	cl.olock.Unlock()
	go func() {
		for {
			fr, ok := <-inchan
			if !ok {
				close(outchan)
				return
			}
			outchan <- fr
			finished, ok := fr.GetFirstHeader("finished")
			if ok && finished == "true" {
				close(outchan)
				cl.closeSeqno(fr.SeqNo)
				return
			}
		}
	}()
	return outchan
}
func (cl *BW2Client) closeSeqno(seqno int) {
	cl.olock.Lock()
	ch, ok := cl.seqnos[seqno]
	if ok {
		close(ch)
		delete(cl.seqnos, seqno)
	}
	cl.olock.Unlock()
}
