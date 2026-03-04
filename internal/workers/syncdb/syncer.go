package syncdb

import "context"

type syncer interface {
	Start(context.Context) error
}
