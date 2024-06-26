package link

import (
	"errors"
	fmt "fmt"

	"github.com/tnnmigga/core/codec"
	"github.com/tnnmigga/core/conc"
	"github.com/tnnmigga/core/conf"
	"github.com/tnnmigga/core/idef"
	"github.com/tnnmigga/core/infra/zlog"
	"github.com/tnnmigga/core/msgbus"

	"github.com/nats-io/nats.go"
)

func (m *module) initHandler() {
	msgbus.RegisterHandler(m, m.onCastPackage)
	msgbus.RegisterHandler(m, m.onStreamCastPackage)
	msgbus.RegisterHandler(m, m.onBroadcastPackage)
	msgbus.RegisterHandler(m, m.onRandomCastPackage)
	msgbus.RegisterHandler(m, m.onRPContext)
}

func (m *module) onCastPackage(pkg *idef.CastPackage) {
	b := codec.Encode(pkg.Body)
	err := m.conn.Publish(castSubject(pkg.ServerID), b)
	if err != nil {
		zlog.Errorf("onCastPackage error %v", err)
	}
}

func (m *module) onStreamCastPackage(pkg *idef.StreamCastPackage) {
	b := codec.Encode(pkg.Body)
	msg := &nats.Msg{
		Subject: streamCastSubject(pkg.ServerID),
		Data:    b,
	}
	if len(pkg.Header) > 0 {
		msg.Header = nats.Header{}
		for key, value := range pkg.Header {
			msg.Header.Set(key, value)
		}
	}
	_, err := m.js.PublishMsgAsync(msg)
	if err != nil {
		zlog.Errorf("onStreamCastPackage error %v", err)
	}
}

func (m *module) onBroadcastPackage(pkg *idef.BroadcastPackage) {
	b := codec.Encode(pkg.Body)
	err := m.conn.Publish(broadcastSubject(pkg.ServerType), b)
	if err != nil {
		zlog.Errorf("onBroadcastPackage error %v", err)
	}
}

func (m *module) onRandomCastPackage(pkg *idef.RandomCastPackage) {
	b := codec.Encode(pkg.Body)
	err := m.conn.Publish(randomCastSubject(pkg.ServerType), b)
	if err != nil {
		zlog.Errorf("onRandomCastPackage error %v", err)
	}
}

func (m *module) onRPContext(ctx *idef.RPCContext) {
	b := codec.Encode(ctx.Req)
	conc.Go(func() {
		resp := &idef.RPCResponse{
			Module: ctx.Caller,
			Req:    ctx.Req,
			Cb:     ctx.Cb,
			Resp:   ctx.Resp,
		}
		defer ctx.Caller.Assign(resp)
		var subject string
		if ctx.ServerType != "" {
			subject = randomRpcSubject(ctx.ServerType)
		} else if ctx.ServerID != 0 {
			subject = rpcSubject(ctx.ServerID)
		} else {
			resp.Err = errors.New("invalid rpc context")
			return
		}
		msg, err := m.conn.Request(subject, b, conf.MaxRPCWaitTime)
		if err != nil {
			resp.Err = err
			return
		}
		data, err := codec.Decode(msg.Data)
		if err != nil {
			resp.Err = fmt.Errorf("RPCPkg decode error: %v", err)
			return
		}
		rpcResp := data.(*RPCResult)
		if len(rpcResp.Err) != 0 {
			resp.Err = errors.New(rpcResp.Err)
			return
		}
		resp.Err = codec.Unmarshal(rpcResp.Data, resp.Resp)
	})
}
