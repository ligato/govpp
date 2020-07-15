// Code generated by GoVPP's binapi-generator. DO NOT EDIT.

package vmxnet3

import (
	"context"
	"fmt"
	api "git.fd.io/govpp.git/api"
	vpe "git.fd.io/govpp.git/binapi/vpe"
	"io"
)

// RPCService defines RPC service for VPP binary API vmxnet3.
type RPCService interface {
	Vmxnet3Create(ctx context.Context, in *Vmxnet3Create) (*Vmxnet3CreateReply, error)
	Vmxnet3Delete(ctx context.Context, in *Vmxnet3Delete) (*Vmxnet3DeleteReply, error)
	Vmxnet3Dump(ctx context.Context, in *Vmxnet3Dump) (RPCService_Vmxnet3DumpClient, error)
}

type serviceClient struct {
	conn api.Connection
}

func NewServiceClient(conn api.Connection) RPCService {
	return &serviceClient{conn}
}

func (c *serviceClient) Vmxnet3Create(ctx context.Context, in *Vmxnet3Create) (*Vmxnet3CreateReply, error) {
	out := new(Vmxnet3CreateReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) Vmxnet3Delete(ctx context.Context, in *Vmxnet3Delete) (*Vmxnet3DeleteReply, error) {
	out := new(Vmxnet3DeleteReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) Vmxnet3Dump(ctx context.Context, in *Vmxnet3Dump) (RPCService_Vmxnet3DumpClient, error) {
	stream, err := c.conn.NewStream(ctx)
	if err != nil {
		return nil, err
	}
	x := &serviceClient_Vmxnet3DumpClient{stream}
	if err := x.Stream.SendMsg(in); err != nil {
		return nil, err
	}
	if err = x.Stream.SendMsg(&vpe.ControlPing{}); err != nil {
		return nil, err
	}
	return x, nil
}

type RPCService_Vmxnet3DumpClient interface {
	Recv() (*Vmxnet3Details, error)
	api.Stream
}

type serviceClient_Vmxnet3DumpClient struct {
	api.Stream
}

func (c *serviceClient_Vmxnet3DumpClient) Recv() (*Vmxnet3Details, error) {
	msg, err := c.Stream.RecvMsg()
	if err != nil {
		return nil, err
	}
	switch m := msg.(type) {
	case *Vmxnet3Details:
		return m, nil
	case *vpe.ControlPingReply:
		return nil, io.EOF
	default:
		return nil, fmt.Errorf("unexpected message: %T %v", m, m)
	}
}
