package updatebonusparentmemberv1

import (
	"context"
)

type Inport interface {
	Execute(ctx context.Context, memberID string, passParent bool, isAdded bool) (bool, error)
}
