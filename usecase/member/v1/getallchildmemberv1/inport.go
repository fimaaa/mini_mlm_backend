package getallchildmemberv1

import (
	"backend_base_app/domain/entity"
	"context"
)

type Inport interface {
	Execute(ctx context.Context, id string, maxLevel int) ([]entity.MemberTree, error)
}
