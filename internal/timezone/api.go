package timezone

import "context"

type ApiTimezone interface {
	Lookup(ctx context.Context, lat, lon float64) (ianaTZ string, err error)
}
