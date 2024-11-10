package deleteonememberv1

import (
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

func (r *apibaseappmembercreateInteractor) Execute(ctx context.Context, id string) (bool, error) {
	res := false

	err := dbhelpers.WithoutTransaction(ctx, r.outport, func(ctx context.Context) error {
		organizerWeddingDataData, err := r.outport.DeleteOneMemberData(ctx, id)
		if err != nil {
			return err
		}

		res = organizerWeddingDataData

		return nil
	})

	return res, err
}
