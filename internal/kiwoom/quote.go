package kiwoom

import (
	"context"
	"strings"

	kiwoomspecs "github.com/smallfish06/krsec/internal/kiwoom/specs"
	"github.com/smallfish06/krsec/pkg/broker"
)

// InquirePriceByRequest fetches ka10001 and returns typed quote fields.
func (c *Client) InquirePriceByRequest(
	ctx context.Context,
	req kiwoomspecs.KiwoomApiDostkStkinfoKa10001Request,
) (*kiwoomspecs.KiwoomApiDostkStkinfoKa10001Response, error) {
	req.StkCd = normalizeSymbolCode(req.StkCd)
	if req.StkCd == "" {
		return nil, broker.ErrInvalidSymbol
	}

	respObj, err := c.CallDocumentedEndpoint(ctx, "ka10001", PathStockInfo, &req)
	if err != nil {
		return nil, err
	}
	out := &kiwoomspecs.KiwoomApiDostkStkinfoKa10001Response{}
	if err := bindResponseObject(respObj, out); err != nil {
		return nil, err
	}
	if strings.TrimSpace(out.StkCd) == "" {
		out.StkCd = req.StkCd
	}
	return out, nil
}

// InquirePrice fetches ka10001 and returns typed quote fields.
func (c *Client) InquirePrice(ctx context.Context, symbol string) (*kiwoomspecs.KiwoomApiDostkStkinfoKa10001Response, error) {
	return c.InquirePriceByRequest(ctx, kiwoomspecs.KiwoomApiDostkStkinfoKa10001Request{
		StkCd: symbol,
	})
}

func setDefaultString(target *string, value string) {
	if target == nil {
		return
	}
	if strings.TrimSpace(*target) == "" {
		*target = value
	}
}
