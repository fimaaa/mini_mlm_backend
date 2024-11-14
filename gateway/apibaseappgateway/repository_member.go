package apibaseappgateway

import (
	"backend_base_app/domain/domerror"
	"backend_base_app/domain/entity"
	"backend_base_app/gateway"
	"backend_base_app/shared/dbhelpers"
	"backend_base_app/shared/log"
	"backend_base_app/shared/util"
	"fmt"
	"time"

	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CreateMemberDataRepo interface {
	CreateMemberData(ctx context.Context, obj entity.MemberData) error
	FindOneMemberDataById(ctx context.Context, id string, skipCalculateBonus bool) (*entity.MemberDataShown, error)
	DeleteOneMemberData(ctx context.Context, id string) (bool, error)
	UpdateMemberData(ctx context.Context, memberData entity.MemberDataShown) (*entity.MemberDataShown, error)
	FindAllMemberData(ctx context.Context, req entity.BaseReqFind) ([]*entity.MemberListShown, int64, error)
	MemberLoginAuthorization(ctx context.Context, obj entity.MemberReqAuth) (*entity.MemberDataShown, error)
	UpdateMemberManualData(ctx context.Context, memberData entity.EditMemberData) (*entity.MemberDataShown, error)
	UpdateBonusMemberData(ctx context.Context, memberID string, passParent bool, isAdded bool, oldParentId *string) (bool, error)
	FindTotalMemberWithParent(ctx context.Context, parentId string) (int64, error)
	GetAllChildMember(ctx context.Context, id string, maxLevel int) ([]entity.MemberTree, error)
}

type memberCollection struct {
	*mongo.Collection
}

func (r GatewayApiBaseApp) getMemberCollection() memberCollection {
	return memberCollection{
		r.MongoWithTransactionImpl.MongoClient.Database(r.database).Collection(entity.CollectionMember),
	}
}

func getFilterMemberKeyword(
	obj entity.MemberDataFind,
	onlySimiliar bool,
) primitive.M {
	//====== execute query using transaction ======
	//count the existing users
	keywordFilter := make([]bson.M, 0)

	if obj.Username != "" {
		keyword := bson.M{"username": obj.Username}
		if onlySimiliar {
			keyword = bson.M{"username": primitive.Regex{Pattern: string(obj.Username), Options: "i"}}
		}
		keywordFilter = append(keywordFilter, keyword)
	}
	if obj.Fullname != "" {
		keyword := bson.M{"fullname": obj.Fullname}
		if onlySimiliar {
			keyword = bson.M{"fullname": primitive.Regex{Pattern: string(obj.Fullname), Options: "i"}}
		}
		keywordFilter = append(keywordFilter, keyword)
	}
	if obj.MemberType != "" {
		keyword := bson.M{"member_type": obj.MemberType}
		keywordFilter = append(keywordFilter, keyword)
	}
	if obj.IsSuspend != nil {
		keyword := bson.M{"is_suspend": obj.IsSuspend}
		keywordFilter = append(keywordFilter, keyword)
	}
	if obj.CreatedAtFrom != nil && !obj.CreatedAtFrom.IsZero() {
		createdAtTo := time.Now()
		if obj.CreatedAtTo != nil && !obj.CreatedAtTo.IsZero() {
			createdAtTo = *obj.CreatedAtTo
		}
		keyword := bson.M{
			"created_at": bson.M{
				"$gte": obj.CreatedAtFrom, // Greater than or equal to date_from
				"$lte": createdAtTo,       // Less than or equal to date_to
			},
		}
		// Append the date filter to the keywordFilter slice
		keywordFilter = append(keywordFilter, keyword)
	}
	if obj.UpdatedAtFrom != nil && !obj.UpdatedAtFrom.IsZero() {
		updatedAtTo := time.Now()
		if obj.UpdatedAtTo != nil && !obj.UpdatedAtTo.IsZero() {
			updatedAtTo = *obj.UpdatedAtTo
		}
		keyword := bson.M{
			"updated_at": bson.M{
				"$gte": obj.CreatedAtFrom, // Greater than or equal to date_from
				"$lte": updatedAtTo,       // Less than or equal to date_to
			},
		}

		// Append the date filter to the keywordFilter slice
		keywordFilter = append(keywordFilter, keyword)
	}
	if obj.LastLoginFrom != nil && !obj.LastLoginFrom.IsZero() {
		lastLoginTo := time.Now()
		if obj.LastLoginTo != nil && !obj.LastLoginTo.IsZero() {
			lastLoginTo = *obj.LastLoginTo
		}
		keyword := bson.M{
			"last_login": bson.M{
				"$gte": obj.LastLoginFrom, // Greater than or equal to date_from
				"$lte": lastLoginTo,       // Less than or equal to date_to
			},
		}

		// Append the date filter to the keywordFilter slice
		keywordFilter = append(keywordFilter, keyword)
	}
	if obj.PhoneNumber != "" {
		keyword := bson.M{"phone_number": obj.PhoneNumber}
		if onlySimiliar {
			keyword = bson.M{"phone_number": primitive.Regex{Pattern: string(obj.PhoneNumber), Options: "i"}}
		}
		keywordFilter = append(keywordFilter, keyword)
	}
	if obj.Email != "" {
		keyword := bson.M{"email": obj.Email}
		if onlySimiliar {
			keyword = bson.M{"email": primitive.Regex{Pattern: string(obj.Email), Options: "i"}}
		}
		keywordFilter = append(keywordFilter, keyword)
	}

	if obj.WithParent {
		keyword := bson.M{"parent_id": obj.ParentId}
		keywordFilter = append(keywordFilter, keyword)
	}

	var allCriteria []bson.M
	var criteriaKeyword bson.M
	if len(keywordFilter) > 0 {
		criteriaKeyword = bson.M{"$or": keywordFilter}
		allCriteria = append(allCriteria, criteriaKeyword)
	}

	if len(obj.BannedId) > 0 {
		exclusionFilter := bson.M{
			"$and": bson.A{
				bson.M{"_id": bson.M{"$nin": obj.BannedId}},       // Exclude documents with _id in bannedIDs
				bson.M{"parent_id": bson.M{"$nin": obj.BannedId}}, // Exclude documents with parent_id in bannedIDs
			},
		}
		fmt.Println("LEN exclusionFilter ", exclusionFilter)
		allCriteria = append(allCriteria, exclusionFilter)
	}

	criteria := bson.M{}
	if len(allCriteria) > 0 {
		criteria = bson.M{"$and": allCriteria}
	}
	return criteria
}

func (coll memberCollection) GetTotalMember(ctx context.Context, obj entity.MemberDataFind, onlySimiliar bool) (int64, error) {
	criteria := getFilterMemberKeyword(obj, onlySimiliar)

	countOpts := options.CountOptions{}
	return coll.CountDocuments(ctx, criteria, &countOpts)
}

func (r GatewayApiBaseApp) CreateMemberData(ctx context.Context, obj entity.MemberData) error {
	log.Info(ctx, "called")

	var err error

	memberCollection := r.getMemberCollection()

	count, err := memberCollection.GetTotalMember(
		ctx,
		entity.MemberDataFind{
			Username:    obj.Username,
			PhoneNumber: obj.PhoneNumber,
			Email:       obj.Email,
		},
		false,
	)

	if err != nil {
		log.Error(ctx, err.Error())
		return err
	}
	if count > 0 {
		return error(DataRegistraionHasTaken)
	}

	err = dbhelpers.WithTransaction(ctx, r.MongoWithTransactionImpl, func(dbCtx context.Context) error {
		info, err := memberCollection.InsertOne(ctx, obj)
		log.Info(ctx, "info >>> ", info)
		if err != nil {
			return err
		}
		fmt.Println("Created Successfuly")

		insertedID := ""
		if oid, ok := info.InsertedID.(string); ok {
			insertedID = oid
		} else {
			// Handle other possible types of ID here if needed
			fmt.Println("Inserted ID is not of type ObjectID")
			return err
		}
		fmt.Println("Inserted ID ", insertedID)

		_, err = r.UpdateBonusMemberData(ctx, insertedID, true, true, nil)
		if err != nil {
			return err
		}
		return err
	})

	return err
}

func (r GatewayApiBaseApp) FindOneMemberDataById(ctx context.Context, id string, skipCalculateBonus bool) (*entity.MemberDataShown, error) {
	log.Info(ctx, "called")

	var (
		resultMemberData entity.MemberDataShown
		err              error
	)

	coll := r.getMemberCollection()
	resCol := coll.FindOne(ctx, bson.M{"_id": id})
	err = resCol.Decode(&resultMemberData)
	if err != nil {
		log.Error(ctx, "id"+id+" "+err.Error())

		// error return if data not found
		if err.Error() == "mongo: no documents in result" {
			return nil, fmt.Errorf("member data not found")
		}

		return nil, err
	}

	if skipCalculateBonus {
		return &resultMemberData, nil
	}

	newBonus, bonusErr := r.CalculateBonusMember(ctx, id, float64(resultMemberData.Bonus))
	fmt.Println("TAG newBonus ", newBonus, " == ", resultMemberData)
	if bonusErr == nil && newBonus != float64(resultMemberData.Bonus) {
		newResultMemberData, err := r.UpdateMemberManualData(ctx, entity.EditMemberData{
			ID:       id,
			Bonus:    float32(newBonus),
			ParentId: resultMemberData.ParentId,
		})
		if err == nil {
			resultMemberData = *newResultMemberData
		}
	}

	return &resultMemberData, nil
}

func (r GatewayApiBaseApp) UpdateMemberData(ctx context.Context, memberData entity.MemberDataShown) (*entity.MemberDataShown, error) {

	log.Info(ctx, "called ", memberData.ID)

	memberData.UpdatedAt = time.Now().Local().UTC()

	coll := r.getMemberCollection()
	// fmt.Println("TAG REPO MEMBER UPDATE ", memberData)

	oldMember, err := r.FindOneMemberDataById(ctx, memberData.ID, true)
	if err != nil {
		return nil, err
	}

	r.UpdateBonusMemberData(ctx, memberData.ID, true, false, oldMember.ParentId)

	dataMap := bson.M{}
	bsonBytes, err := bson.Marshal(memberData)
	if err != nil {
		fmt.Errorf("failed to marshal struct: %w", err)
		return &memberData, err
	}
	err = bson.Unmarshal(bsonBytes, &dataMap)
	if err != nil {
		fmt.Errorf("failed to unmarshal to map: %w", err)
		return &memberData, err
	}

	// Remove `_id` from the map to avoid updating the immutable field
	delete(dataMap, "_id")

	// Define the filter by ID
	filter := bson.M{"_id": memberData.ID}

	// Define the update operation
	update := bson.M{"$set": dataMap}

	// Perform the update operation
	_, err = coll.UpdateOne(ctx, filter, update)

	if err != nil {
		log.Info(ctx, "error >>> "+err.Error())
	}

	r.UpdateBonusMemberData(ctx, memberData.ID, true, true, nil)

	return &memberData, err
}

func (r GatewayApiBaseApp) FindAllMemberData(ctx context.Context, req entity.BaseReqFind) ([]*entity.MemberListShown, int64, error) {
	log.Info(ctx, "called")

	var (
		err  error
		objs []*entity.MemberListShown
	)

	coll := r.getMemberCollection()

	var findData entity.MemberDataFind
	if err := util.MapToStruct(req.Value, &findData); err != nil {
		fmt.Println("Error:", err)
		return nil, 0, err
	}
	fmt.Println("TAG 329 findData ", req.Value, " -- ", findData)

	criteria := getFilterMemberKeyword(findData, true)

	fmt.Println("TAG 332 findData ", criteria)

	findOpts := gateway.BaseReqFindToOptOption(req)

	cursor, err := coll.Find(ctx, criteria, &findOpts)
	if err != nil {
		fmt.Println("TAG ALL MEMBER error 1 ", err)
		return nil, 0, err
	}

	if err := cursor.All(ctx, &objs); err != nil {
		fmt.Println("TAG ALL MEMBER error 2 ", err)
		return nil, 0, err
	}

	//counting
	count, err := coll.GetTotalMember(ctx, findData, true)

	return objs, count, err
}

func (r GatewayApiBaseApp) MemberLoginAuthorization(ctx context.Context, obj entity.MemberReqAuth) (*entity.MemberDataShown, error) {
	log.Info(ctx, "called")

	var (
		resultMemberDataShown *entity.MemberDataShown
		err                   error
	)

	coll := r.getMemberCollection()

	encryptPassword := r.EncryptPassword(ctx, obj.Password)
	// encryptPassword := obj.Password

	log.Info(ctx, "TAG PASSWORD ", encryptPassword)

	err = coll.FindOne(ctx, bson.M{"username": obj.Username, "password": encryptPassword}).Decode(&resultMemberDataShown)

	if err != nil {
		return resultMemberDataShown, err
	}

	if resultMemberDataShown.IsSuspend {
		err = entity.NewMyError("Account is Suspended")
		return resultMemberDataShown, err
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")
	resultMemberDataShown.LastLogin = time.Now().In(loc)

	if obj.DeviceId != "" {
		resultMemberDataShown.DeviceId = obj.DeviceId
	}
	if obj.TokenBroadcast != "" {
		resultMemberDataShown.TokenBroadcast = obj.TokenBroadcast
	}

	fmt.Println("TAG REPO MEMbER LOGIN OBJ ", obj)

	fmt.Println("TAG REPO MEMbER LOGIN resultMemberDataShown ", resultMemberDataShown)

	return r.UpdateMemberData(ctx, *resultMemberDataShown)
}

func (r GatewayApiBaseApp) DeleteOneMemberData(ctx context.Context, id string) (bool, error) {
	var (
		err error
	)

	coll := r.getMemberCollection()
	r.UpdateBonusMemberData(ctx, id, true, false, nil)

	resCol, err := coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		log.Error(ctx, err.Error())

		// error return if data not found
		if err.Error() == "mongo: no documents in result" {
			err = fmt.Errorf("member data not found")
		}

		return false, err
	}

	return resCol.DeletedCount > 0, err
}

func (r GatewayApiBaseApp) UpdateMemberManualData(ctx context.Context, memberData entity.EditMemberData) (*entity.MemberDataShown, error) {
	log.Info(ctx, "called ", memberData.ID)

	coll := r.getMemberCollection()
	var dataResult entity.MemberDataShown

	oldMember, err := r.FindOneMemberDataById(ctx, memberData.ID, true)
	if err != nil {
		return nil, err
	}

	if memberData.ParentId != oldMember.ParentId {
		newLevel := int32(0)
		if memberData.ParentId != nil {
			parentNewMember, err := r.FindOneMemberDataById(ctx, *memberData.ParentId, true)
			if err != nil {
				return nil, err
			}
			if parentNewMember.Level != nil {
				newLevel = 1 + *parentNewMember.Level
			}
		}
		memberData.Level = &newLevel
		r.UpdateBonusMemberData(ctx, memberData.ID, true, false, oldMember.ParentId)
	}

	memberData.UpdatedAt = time.Now().Local().UTC()

	log.Info(ctx, "EDIT MEMBERDATA ", memberData)

	editBson := util.StructToBSONM(memberData)

	log.Info(ctx, "EDIT BSON ", editBson)
	// Prepare the update document with $set operator
	update := bson.M{
		"$set": editBson,
	}

	// Use FindOneAndUpdate with filter on member ID
	err = coll.FindOneAndUpdate(ctx, bson.M{"_id": memberData.ID}, update).Decode(&dataResult)

	if err != nil {
		log.Info(ctx, "error >>> "+err.Error())
		return nil, err
	}

	if memberData.ParentId != oldMember.ParentId {
		r.UpdateBonusMemberData(ctx, memberData.ID, true, true, nil)
	}

	return &dataResult, err
}

func (r GatewayApiBaseApp) UpdateBonusMemberData(ctx context.Context, memberID string, passParent bool, isAdded bool, oldParentId *string) (bool, error) {
	parentID := oldParentId
	isSucess := false

	fmt.Println("Update Bonus ", memberID, " == ", parentID)
	if parentID == nil {
		memberChild, err := r.FindOneMemberDataById(ctx, memberID, true)
		if err != nil {
			fmt.Println("MemberId not found ", err)
			return isSucess, err
		}

		fmt.Println("MemberId found ", memberID)
		if memberChild.ParentId == nil {
			isSucess = true
			return isSucess, nil
		}
		parentID = memberChild.ParentId
	}

	memberParent, err := r.FindOneMemberDataById(ctx, *parentID, true)
	if err != nil {
		fmt.Println("MemberParentId not found ", err)
		return isSucess, nil
	}
	fmt.Println("MemberParentId found ", parentID)

	changeValue := 0.5
	if passParent {
		changeValue = 1
	} else {
		childBonus, err := r.FindTotalMemberWithParent(ctx, *parentID)
		if err == nil {
			fmt.Println("childBonus not found ", err)
			childBonusValue := float64(childBonus) * 0.5
			changeValue = changeValue + childBonusValue
		}
	}
	if !isAdded {
		changeValue = -changeValue
	}

	_, err = r.UpdateMemberManualData(ctx, entity.EditMemberData{
		ID:       *parentID,
		Bonus:    memberParent.Bonus + float32(changeValue),
		ParentId: memberParent.ParentId,
	})

	if err != nil {
		fmt.Println("UpdateMemberManualData not found ", err)
		return isSucess, err
	}

	fmt.Println("UpdateMemberManualData found ", err)
	if !passParent || parentID == nil {
		isSucess = true
		return isSucess, err
	}

	return r.UpdateBonusMemberData(ctx, *parentID, false, isAdded, nil)
}

func (r GatewayApiBaseApp) CalculateBonusMember(ctx context.Context, id string, oldBonus float64) (float64, error) {
	dataFind := entity.MemberDataFind{
		ParentId:   &id,
		WithParent: true,
	}
	dataMap, err := util.StructToMap(dataFind)
	if err != nil {
		return oldBonus, err
	}
	listFirstChildMember, count, err := r.FindAllMemberData(ctx, entity.BaseReqFind{
		Size:  0,
		Page:  0,
		Value: dataMap,
	})
	if count <= 0 {
		return 0, err
	}

	bonusFirstChild := count * 1.0
	bonusSecondChild := 0.0
	for _, member := range listFirstChildMember {
		fmt.Println("TAG memberID ", member.ID)
		totalChild, err := r.FindTotalMemberWithParent(ctx, member.ID)
		if err == nil {
			bonusSecondChild = bonusSecondChild + (float64(totalChild) * 0.5)
		}
		fmt.Println("TAG TOTALCHILD ", member.Username, " == ", totalChild, " => ", bonusSecondChild)
	}

	return (float64(bonusFirstChild) + bonusSecondChild), err
}

func (r GatewayApiBaseApp) GetAllChildMember(ctx context.Context, id string, maxLevel int) ([]entity.MemberTree, error) {
	var objs []entity.MemberTree
	coll := r.getMemberCollection()

	cursor, err := coll.Find(ctx, bson.M{"parent_id": id})
	if err != nil {
		fmt.Println("TAG ALL MEMBER error 1 ", err)
		return nil, err
	}

	if err := cursor.All(ctx, &objs); err != nil {
		fmt.Println("TAG ALL MEMBER error 2 ", err)
		return nil, err
	}

	if len(objs) <= 0 {
		return objs, err
	}

	fmt.Println("TAG MAXLEVEL ", maxLevel, " <= ", *objs[0].Level)
	if objs[0].Level == nil || *objs[0].Level > int32(maxLevel) {
		return objs, err
	}

	for i, obj := range objs {
		fmt.Println("TAG MAXCHILDREN ", maxLevel, " <= ", obj.ID)
		child, _ := r.GetAllChildMember(ctx, obj.ID, maxLevel)
		objs[i].ChildMember = &child
		fmt.Println("TAG obj.children ", child, " ==> ", obj.ChildMember)
	}

	for _, obj := range objs {
		fmt.Println("TAG nuxt.children ", obj.Username, " ==> ", obj.ChildMember)
	}

	fmt.Println("TAG RESPONSE ", objs)
	return objs, err
}

func (r GatewayApiBaseApp) FindTotalMemberWithParent(ctx context.Context, parentId string) (int64, error) {
	coll := r.getMemberCollection()
	return coll.GetTotalMember(ctx, entity.MemberDataFind{
		WithParent: true,
		ParentId:   &parentId,
	}, false)
}

const DataRegistraionHasTaken domerror.ErrorType = "ER1006 data registration has been taken"
