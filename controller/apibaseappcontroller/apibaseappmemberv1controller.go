package apibaseappcontroller

import (
	"backend_base_app/domain/domerror"
	"backend_base_app/domain/entity"
	"backend_base_app/shared/log"
	"backend_base_app/shared/util"
	"backend_base_app/usecase/member/v1/creatememberv1"
	"backend_base_app/usecase/member/v1/deleteonememberv1"
	"backend_base_app/usecase/member/v1/editmemberv1"
	"backend_base_app/usecase/member/v1/getallchildmemberv1"
	"backend_base_app/usecase/member/v1/getallmemberv1"
	"backend_base_app/usecase/member/v1/getmemberv1"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func ApiBaseAppMemberCreate(r *Controller) gin.HandlerFunc {
	var inputPort = creatememberv1.NewUsecase(r.DataSource)

	return func(c *gin.Context) {
		traceID := util.GenerateID()
		ctx := log.Context(c.Request.Context(), traceID)

		var req entity.CreateMemberData
		if err := c.Bind(&req); err != nil {
			newErr := domerror.FailUnmarshalRequestBodyError
			log.Error(ctx, err.Error())
			r.Helper.SendBadRequest(c, err.Error(), newErr, traceID)
			return
		}

		if err := req.ValidateCreate(); err != nil {
			r.Helper.SendBadRequest(c, err.Error(), nil, traceID)
			return
		}

		res, err := inputPort.Execute(ctx, req)

		if err != nil {
			log.Error(ctx, err.Error())
			r.Helper.SendBadRequest(c, err.Error(), fmt.Sprintf("file err : %s", err.Error()), traceID)
			return
		}

		r.Helper.SendSuccess(c, "Success", res, traceID, nil)
	}
}

func ApiBaseAppMemberFindAll(r *Controller) gin.HandlerFunc {
	var inputPort = getallmemberv1.NewUsecase(r.DataSource)

	return func(c *gin.Context) {
		traceID := util.GenerateID()
		ctx := log.Context(c.Request.Context(), traceID)

		var req entity.BaseReqFind
		if err := c.BindQuery(&req); err != nil {
			newErr := domerror.FailUnmarshalRequestBodyError
			log.Error(ctx, err.Error())
			r.Helper.SendBadRequest(c, err.Error(), newErr, traceID)
			return
		}

		sortByParams := make(map[string]interface{})
		// Loop through all query parameters
		for key, value := range c.Request.URL.Query() {
			// Check if the key contains "sort_by_"
			if strings.HasPrefix(key, "sort_by_") {
				trimmedKey := strings.TrimPrefix(key, "sort_by_")
				sortByParams[trimmedKey] = value
			}
		}
		req.SortBy = sortByParams

		res, count, err := inputPort.Execute(ctx, req)

		if err != nil {
			log.Error(ctx, err.Error())
			r.Helper.SendBadRequest(c, err.Error(), fmt.Sprintf("file err : %s", err.Error()), traceID)
			return
		}

		finalResponse := req.ToResponse(res, count)

		r.Helper.SendSuccess(c, "Success", finalResponse.List, traceID, finalResponse.Pagination)
	}
}

func ApiBaseAppMemberFindOne(r *Controller) gin.HandlerFunc {
	var inputPort = getmemberv1.NewUsecase(r.DataSource)

	return func(c *gin.Context) {
		traceID := util.GenerateID()
		ctx := log.Context(c.Request.Context(), traceID)

		var req = c.Param("id")
		if req == "" {
			newErr := domerror.FailUnmarshalRequestBodyError
			r.Helper.SendBadRequest(c, string(newErr), newErr, traceID)
			return
		}

		fmt.Println("FIND MEMBER ID ", req)
		res, err := inputPort.Execute(ctx, req)

		if err != nil {
			log.Error(ctx, err.Error())
			r.Helper.SendBadRequest(c, err.Error(), fmt.Sprintf("file err : %s", err.Error()), traceID)
			return
		}

		r.Helper.SendSuccess(c, "Success", res, traceID, nil)
	}
}

func ApiBaseAppMemberUpdateOne(r *Controller) gin.HandlerFunc {
	var inputPort = editmemberv1.NewUsecase(r.DataSource)

	return func(c *gin.Context) {
		traceID := util.GenerateID()
		ctx := log.Context(c.Request.Context(), traceID)

		var req entity.EditMemberData
		if err := c.Bind(&req); err != nil {
			newErr := domerror.FailUnmarshalRequestBodyError
			log.Error(ctx, err.Error())
			r.Helper.SendBadRequest(c, err.Error(), newErr, traceID)
			return
		}

		fmt.Println("FIND MEMBER ID ", req)
		res, err := inputPort.Execute(ctx, req)

		if err != nil {
			log.Error(ctx, err.Error())
			r.Helper.SendBadRequest(c, err.Error(), fmt.Sprintf("file err : %s", err.Error()), traceID)
			return
		}

		r.Helper.SendSuccess(c, "Success", res, traceID, nil)
	}
}

func ApiBaseAppMemberDeleteOne(r *Controller) gin.HandlerFunc {
	var inputPort = deleteonememberv1.NewUsecase(r.DataSource)

	return func(c *gin.Context) {
		traceID := util.GenerateID()
		ctx := log.Context(c.Request.Context(), traceID)

		var req = c.Param("id")
		if req == "" {
			newErr := domerror.FailUnmarshalRequestBodyError
			r.Helper.SendBadRequest(c, string(newErr), newErr, traceID)
			return
		}

		res, err := inputPort.Execute(ctx, req)

		if err != nil {
			log.Error(ctx, err.Error())
			r.Helper.SendBadRequest(c, err.Error(), fmt.Sprintf("file err : %s", err.Error()), traceID)
			return
		}

		r.Helper.SendSuccess(c, "Success", res, traceID, nil)
	}
}

func ApiBaseAppMemberFindAllChildren(r *Controller) gin.HandlerFunc {
	var inputPort = getallchildmemberv1.NewUsecase(r.DataSource)

	return func(c *gin.Context) {
		traceID := util.GenerateID()
		ctx := log.Context(c.Request.Context(), traceID)

		id := c.Param("id")
		maxLevelStr := c.DefaultQuery("maxLevel", "0")

		// Convert maxLevel from string to int
		maxLevel, err := strconv.Atoi(maxLevelStr)
		if err != nil {
			// Handle the case where maxLevel is not a valid integer
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid maxLevel"})
			return
		}

		res, err := inputPort.Execute(ctx, id, maxLevel)

		if err != nil {
			log.Error(ctx, err.Error())
			r.Helper.SendBadRequest(c, err.Error(), fmt.Sprintf("file err : %s", err.Error()), traceID)
			return
		}

		r.Helper.SendSuccess(c, "Success", res, traceID, nil)
	}
}
