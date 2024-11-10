package gateway

import (
	"backend_base_app/domain/entity"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func BaseReqFindToOptOption(req entity.BaseReqFind) options.FindOptions {
	limitInt64 := int64(req.Size)
	page := 0
	if req.Page > 0 {
		page = req.Page
	}
	skip := int64(req.Size * page)
	var sort []bson.E
	if len(req.SortBy) > 0 {
		for k, v := range req.SortBy {
			sort = append(sort, bson.E{
				Key:   k,
				Value: v,
			})
		}
	} else {
		sort = append(sort, bson.E{
			Key:   "updated_at",
			Value: -1,
		})
	}
	return options.FindOptions{
		Limit: &limitInt64,
		Skip:  &skip,
		Sort:  sort,
	}
}
