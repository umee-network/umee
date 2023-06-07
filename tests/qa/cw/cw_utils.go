package cw

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func GetAttributeValue(resp sdk.TxResponse, eventName, attrKey string) string {
	var attrVal string
	for _, event := range resp.Logs[0].Events {
		if event.Type == eventName {
			for _, attribute := range event.Attributes {
				if attribute.Key == attrKey {
					attrVal = attribute.Value
				}
			}
		}
	}
	return attrVal
}
