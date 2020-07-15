// Code generated by GoVPP's binapi-generator. DO NOT EDIT.

package cdp

import (
	"context"
	api "git.fd.io/govpp.git/api"
)

// RPCService defines RPC service for VPP binary API cdp.
type RPCService interface {
	CdpEnableDisable(ctx context.Context, in *CdpEnableDisable) (*CdpEnableDisableReply, error)
}

type serviceClient struct {
	conn api.Connection
}

func NewServiceClient(conn api.Connection) RPCService {
	return &serviceClient{conn}
}

func (c *serviceClient) CdpEnableDisable(ctx context.Context, in *CdpEnableDisable) (*CdpEnableDisableReply, error) {
	out := new(CdpEnableDisableReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
