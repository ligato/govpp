// Code generated by GoVPP's binapi-generator. DO NOT EDIT.

package lldp

import (
	"context"
	api "git.fd.io/govpp.git/api"
)

// RPCService defines RPC service for VPP binary API lldp.
type RPCService interface {
	LldpConfig(ctx context.Context, in *LldpConfig) (*LldpConfigReply, error)
	SwInterfaceSetLldp(ctx context.Context, in *SwInterfaceSetLldp) (*SwInterfaceSetLldpReply, error)
}

type serviceClient struct {
	conn api.Connection
}

func NewServiceClient(conn api.Connection) RPCService {
	return &serviceClient{conn}
}

func (c *serviceClient) LldpConfig(ctx context.Context, in *LldpConfig) (*LldpConfigReply, error) {
	out := new(LldpConfigReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) SwInterfaceSetLldp(ctx context.Context, in *SwInterfaceSetLldp) (*SwInterfaceSetLldpReply, error) {
	out := new(SwInterfaceSetLldpReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
