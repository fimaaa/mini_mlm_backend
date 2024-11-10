package editmemberv1

import (
	"backend_base_app/domain/entity"
	"backend_base_app/shared/dbhelpers"
	"context"
)

type apibaseappmembercreateInteractor struct {
	outport Outport
}

func NewUsecase(outputPort Outport) Inport {
	return &apibaseappmembercreateInteractor{
		outport: outputPort,
	}
}

func (r *apibaseappmembercreateInteractor) Execute(ctx context.Context, req entity.EditMemberData) (*entity.MemberDataShown, error) {
	res := &entity.MemberDataShown{}

	err := dbhelpers.WithoutTransaction(ctx, r.outport, func(ctx context.Context) error {
		//encrypt password
		if req.Password != nil {
			password := r.outport.EncryptPassword(ctx, *req.Password)
			req.Password = &password
		}

		resUpdate, err := r.outport.UpdateMemberManualData(ctx, req)
		if err != nil {
			return err
		}
		res = resUpdate

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}
