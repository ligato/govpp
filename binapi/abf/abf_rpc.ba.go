// Code generated by GoVPP's binapi-generator. DO NOT EDIT.

package abf

import (
	"context"
	"fmt"
	api "git.fd.io/govpp.git/api"
	vpe "git.fd.io/govpp.git/binapi/vpe"
	"io"
)

// RPCService defines RPC service for VPP binary API abf.
type RPCService interface {
	AbfItfAttachAddDel(ctx context.Context, in *AbfItfAttachAddDel) (*AbfItfAttachAddDelReply, error)
	AbfItfAttachDump(ctx context.Context, in *AbfItfAttachDump) (RPCService_AbfItfAttachDumpClient, error)
	AbfPluginGetVersion(ctx context.Context, in *AbfPluginGetVersion) (*AbfPluginGetVersionReply, error)
	AbfPolicyAddDel(ctx context.Context, in *AbfPolicyAddDel) (*AbfPolicyAddDelReply, error)
	AbfPolicyDump(ctx context.Context, in *AbfPolicyDump) (RPCService_AbfPolicyDumpClient, error)
}

type serviceClient struct {
	conn api.Connection
}

func NewServiceClient(conn api.Connection) RPCService {
	return &serviceClient{conn}
}

func (c *serviceClient) AbfItfAttachAddDel(ctx context.Context, in *AbfItfAttachAddDel) (*AbfItfAttachAddDelReply, error) {
	out := new(AbfItfAttachAddDelReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) AbfItfAttachDump(ctx context.Context, in *AbfItfAttachDump) (RPCService_AbfItfAttachDumpClient, error) {
	stream, err := c.conn.NewStream(ctx)
	if err != nil {
		return nil, err
	}
	x := &serviceClient_AbfItfAttachDumpClient{stream}
	if err := x.Stream.SendMsg(in); err != nil {
		return nil, err
	}
	if err = x.Stream.SendMsg(&vpe.ControlPing{}); err != nil {
		return nil, err
	}
	return x, nil
}

type RPCService_AbfItfAttachDumpClient interface {
	Recv() (*AbfItfAttachDetails, error)
	api.Stream
}

type serviceClient_AbfItfAttachDumpClient struct {
	api.Stream
}

func (c *serviceClient_AbfItfAttachDumpClient) Recv() (*AbfItfAttachDetails, error) {
	msg, err := c.Stream.RecvMsg()
	if err != nil {
		return nil, err
	}
	switch m := msg.(type) {
	case *AbfItfAttachDetails:
		return m, nil
	case *vpe.ControlPingReply:
		return nil, io.EOF
	default:
		return nil, fmt.Errorf("unexpected message: %T %v", m, m)
	}
}

func (c *serviceClient) AbfPluginGetVersion(ctx context.Context, in *AbfPluginGetVersion) (*AbfPluginGetVersionReply, error) {
	out := new(AbfPluginGetVersionReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) AbfPolicyAddDel(ctx context.Context, in *AbfPolicyAddDel) (*AbfPolicyAddDelReply, error) {
	out := new(AbfPolicyAddDelReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) AbfPolicyDump(ctx context.Context, in *AbfPolicyDump) (RPCService_AbfPolicyDumpClient, error) {
	stream, err := c.conn.NewStream(ctx)
	if err != nil {
		return nil, err
	}
	x := &serviceClient_AbfPolicyDumpClient{stream}
	if err := x.Stream.SendMsg(in); err != nil {
		return nil, err
	}
	if err = x.Stream.SendMsg(&vpe.ControlPing{}); err != nil {
		return nil, err
	}
	return x, nil
}

type RPCService_AbfPolicyDumpClient interface {
	Recv() (*AbfPolicyDetails, error)
	api.Stream
}

type serviceClient_AbfPolicyDumpClient struct {
	api.Stream
}

func (c *serviceClient_AbfPolicyDumpClient) Recv() (*AbfPolicyDetails, error) {
	msg, err := c.Stream.RecvMsg()
	if err != nil {
		return nil, err
	}
	switch m := msg.(type) {
	case *AbfPolicyDetails:
		return m, nil
	case *vpe.ControlPingReply:
		return nil, io.EOF
	default:
		return nil, fmt.Errorf("unexpected message: %T %v", m, m)
	}
}
