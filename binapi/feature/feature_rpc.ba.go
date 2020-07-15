// Code generated by GoVPP's binapi-generator. DO NOT EDIT.

package feature

import (
	"context"
	api "git.fd.io/govpp.git/api"
)

// RPCService defines RPC service for VPP binary API feature.
type RPCService interface {
	FeatureEnableDisable(ctx context.Context, in *FeatureEnableDisable) (*FeatureEnableDisableReply, error)
}

type serviceClient struct {
	conn api.Connection
}

func NewServiceClient(conn api.Connection) RPCService {
	return &serviceClient{conn}
}

func (c *serviceClient) FeatureEnableDisable(ctx context.Context, in *FeatureEnableDisable) (*FeatureEnableDisableReply, error) {
	out := new(FeatureEnableDisableReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
