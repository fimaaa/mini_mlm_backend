package updatebonusparentmemberv1

import (
	"backend_base_app/shared/dbhelpers"
	"context"
)

type apibaseappupdatebonusparentmemberInteractor struct {
	outport Outport
}

func NewUsecase(outputPort Outport) Inport {
	return &apibaseappupdatebonusparentmemberInteractor{
		outport: outputPort,
	}
}

func (r *apibaseappupdatebonusparentmemberInteractor) Execute(ctx context.Context, memberID string, passParent bool, isAdded bool) (bool, error) {
	isSucess := false
	err := dbhelpers.WithoutTransaction(ctx, r.outport, func(ctx context.Context) error {
		success, err := r.outport.UpdateBonusMemberData(ctx, memberID, true, true, nil)
		isSucess = success
		return err
	})

	return isSucess, err
}
