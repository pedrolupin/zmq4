// Copyright 2020 The pedrolupin Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package zmq4_test

import (
	"context"
	"sync"
	"testing"

	"github.com/pedrolupin/zmq4"
)

func TestIssue99(t *testing.T) {
	var (
		wg     sync.WaitGroup
		outMsg zmq4.Msg
		inMsg  zmq4.Msg
		ok     = make(chan int)
	)

	ep, err := EndPoint("tcp")
	if err != nil {
		t.Fatalf("could not find endpoint: %+v", err)
	}

	requester := func() {
		defer wg.Done()
		defer close(ok)

		req := zmq4.NewReq(context.Background())
		defer req.Close()

		err := req.Dial(ep)
		if err != nil {
			t.Errorf("could not dial: %+v", err)
			return
		}

		// Test message w/ 3 frames
		outMsg = zmq4.NewMsgFromString([]string{"ZERO", "Hello!", "World!"})
		err = req.Send(outMsg)
		if err != nil {
			t.Errorf("failed to send: %+v", err)
			return
		}

		inMsg, err = req.Recv()
		if err != nil {
			t.Errorf("failed to recv: %+v", err)
			return
		}
	}

	responder := func() {
		defer wg.Done()

		rep := zmq4.NewRep(context.Background())
		defer rep.Close()

		err := rep.Listen(ep)
		if err != nil {
			t.Errorf("could not dial: %+v", err)
			return
		}

		//  Wait for next request from client
		msg, err := rep.Recv()
		if err != nil {
			t.Errorf("could not recv request: %+v", err)
			return
		}

		//  Send reply back to client
		err = rep.Send(msg)
		if err != nil {
			t.Errorf("could not send reply: %+v", err)
			return
		}
		<-ok
	}

	wg.Add(2)

	go requester()
	go responder()

	wg.Wait()

	if want, got := len(outMsg.Frames), len(inMsg.Frames); want != got {
		t.Fatalf("message length mismatch: got=%d, want=%d", got, want)
	}
}
