package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/gogo/protobuf/proto"
)

func PrintOrErr(resp proto.Message, err error, cctx client.Context) error {
	if err != nil {
		return err
	}
	return cctx.PrintProto(resp)
}
