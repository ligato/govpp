// Code generated by GoVPP's binapi-generator. DO NOT EDIT.

package syslog

import (
	"context"
	api "git.fd.io/govpp.git/api"
)

// RPCService defines RPC service for VPP binary API syslog.
type RPCService interface {
	SyslogGetFilter(ctx context.Context, in *SyslogGetFilter) (*SyslogGetFilterReply, error)
	SyslogGetSender(ctx context.Context, in *SyslogGetSender) (*SyslogGetSenderReply, error)
	SyslogSetFilter(ctx context.Context, in *SyslogSetFilter) (*SyslogSetFilterReply, error)
	SyslogSetSender(ctx context.Context, in *SyslogSetSender) (*SyslogSetSenderReply, error)
}

type serviceClient struct {
	conn api.Connection
}

func NewServiceClient(conn api.Connection) RPCService {
	return &serviceClient{conn}
}

func (c *serviceClient) SyslogGetFilter(ctx context.Context, in *SyslogGetFilter) (*SyslogGetFilterReply, error) {
	out := new(SyslogGetFilterReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) SyslogGetSender(ctx context.Context, in *SyslogGetSender) (*SyslogGetSenderReply, error) {
	out := new(SyslogGetSenderReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) SyslogSetFilter(ctx context.Context, in *SyslogSetFilter) (*SyslogSetFilterReply, error) {
	out := new(SyslogSetFilterReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) SyslogSetSender(ctx context.Context, in *SyslogSetSender) (*SyslogSetSenderReply, error) {
	out := new(SyslogSetSenderReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
