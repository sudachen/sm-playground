package client

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	apitypes "github.com/spacemeshos/api/release/go/spacemesh/v1"
	gosmtypes "github.com/spacemeshos/go-spacemesh/common/types"
	"google.golang.org/genproto/googleapis/rpc/status"
)

// GetSmesherId returns the current smesher id configured in the node
func (c *gRPCClient) GetSmesherId() ([]byte, error) {
	s := c.getSmesherServiceClient()
	if resp, err := s.SmesherID(context.Background(), &empty.Empty{}); err != nil {
		return nil, err
	} else {
		return resp.AccountId.Address, nil
	}
}

// IsSmeshing returns true iff the node is currently setup to smesh
func (c *gRPCClient) IsSmeshing() (bool, error) {
	s := c.getSmesherServiceClient()
	if resp, err := s.IsSmeshing(context.Background(), &empty.Empty{}); err != nil {
		return false, err
	} else {
		return resp.IsSmeshing, nil
	}
}

// GetPostStatus returns the current node proof of space status
//func (c *gRPCClient) GetPostStatus() (*apitypes.PostStatus, error) {
///	s := c.getSmesherServiceClient()
//	if resp, err := s.PostStatus(context.Background(), &empty.Empty{}); err != nil {
//		return nil, err
//	} else {
//		return resp.Status, nil
//	}
//}

// StopSmeshing instructs the node to stop smeshing and optionally delete smeshing data file(s)
func (c *gRPCClient) StopSmeshing(deleteFiles bool) (*status.Status, error) {
	s := c.getSmesherServiceClient()
	resp, err := s.StopSmeshing(context.Background(), &apitypes.StopSmeshingRequest{DeleteFiles: deleteFiles})
	if err != nil {
		return nil, err
	}
	return resp.Status, nil
}

// SetRewardsAddress sets the smesher's rewards address
func (c *gRPCClient) SetRewardsAddress(address gosmtypes.Address) (*status.Status, error) {
	s := c.getSmesherServiceClient()
	resp, err := s.SetCoinbase(context.Background(), &apitypes.SetCoinbaseRequest{Id: &apitypes.AccountId{Address: address.Bytes()}})
	if err != nil {
		return nil, err
	}
	return resp.Status, nil
}

// GetRewardsAddress get the smesher's current rewards address
func (c *gRPCClient) GetRewardsAddress() (*gosmtypes.Address, error) {
	s := c.getSmesherServiceClient()
	resp, err := s.Coinbase(context.Background(), &empty.Empty{})
	if err != nil {
		return nil, err
	}
	addr := gosmtypes.BytesToAddress(resp.AccountId.Address)
	return &addr, nil
}
