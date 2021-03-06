// Copyright 2016 CodisLabs. All Rights Reserved.
// Licensed under the MIT (MIT-LICENSE.txt) license.

package proxy

import (
	"sync"

	"github.com/CodisLabs/codis/pkg/proxy/redis"
	"github.com/CodisLabs/codis/pkg/utils/sync2/atomic2"
)

type Request struct {
	Multi []*redis.Resp
	Start int64
	Batch *sync.WaitGroup
	Group *sync.WaitGroup

	OpStr string
	OpFlag
	Broken *atomic2.Bool

	*redis.Resp
	Err error

	Coalesce func() error
}

func (r *Request) IsBroken() bool {
	return r.Broken != nil && r.Broken.Get()
}

func (r *Request) Release() {
	r.Multi = nil
	r.Resp = nil
	r.Coalesce = nil
}

type RequestAlloc struct {
	alloc struct {
		buf []Request
		off int
	}
	batch struct {
		buf []sync.WaitGroup
		off int
	}
}

func (p *RequestAlloc) NewRequest() *Request {
	var d = &p.alloc
	if len(d.buf) == d.off {
		d.buf = make([]Request, 48)
		d.off = 0
	}
	r := &d.buf[d.off]
	d.off += 1
	return r
}

func (p *RequestAlloc) SubRequest(r *Request) *Request {
	x := p.NewRequest()
	x.Start = r.Start
	x.Batch = r.Batch
	x.OpStr = r.OpStr
	x.OpFlag = r.OpFlag
	x.Broken = r.Broken
	return x
}

func (p *RequestAlloc) NewBatch() *sync.WaitGroup {
	var d = &p.batch
	if len(d.buf) == d.off {
		d.buf = make([]sync.WaitGroup, 64)
		d.off = 0
	}
	w := &d.buf[d.off]
	d.off += 1
	return w
}
