package controllers

import (
	"app/base/database"
	"app/manager/middlewares"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdvisoryName string
type SystemID string

type SystemsAdvisoriesRequest struct {
	Systems    []SystemID     `json:"systems"`
	Advisories []AdvisoryName `json:"advisories"`
}

type SystemsAdvisoriesResponse struct {
	Data map[SystemID][]AdvisoryName `json:"data"`
}
type AdvisoriesSystemsResponse struct {
	Data map[AdvisoryName][]SystemID `json:"data"`
}

type systemsAdvisoriesDBLoad struct {
	SystemID   SystemID     `query:"sp.inventory_id" gorm:"column:system_id"`
	AdvisoryID AdvisoryName `query:"am.name" gorm:"column:advisory_id"`
}

var systemsAdvisoriesSelect = database.MustGetSelect(&systemsAdvisoriesDBLoad{})

func systemsAdvisoriesQuery(acc int, systems []SystemID, advisories []AdvisoryName) *gorm.DB {
	query := database.SystemAdvisories(database.Db, acc).
		Select(systemsAdvisoriesSelect).
		Joins("join advisory_metadata am on am.id = sa.advisory_id").
		Order("sp.inventory_id, am.id")
	if len(systems) > 0 {
		query = query.Where("sp.inventory_id::text in (?)", systems)
	}
	if len(advisories) > 0 {
		query = query.Where("am.name in (?)", advisories)
	}
	return query
}

// @Summary View system-advisory pairs for selected systems and advisories
// @Description View system-advisory pairs for selected systems and advisories
// @ID viewSystemsAdvisories
// @Security RhIdentity
// @Accept   json
// @Produce  json
// @Param    body    body    SystemsAdvisoriesRequest true "Request body"
// @Success 200 {object} SystemsAdvisoriesResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /views/systems/advisories [post]
func PostSystemsAdvisories(c *gin.Context) {
	var req SystemsAdvisoriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		LogAndRespBadRequest(c, err, "Invalid request body")
	}
	acc := c.GetInt(middlewares.KeyAccount)
	q := systemsAdvisoriesQuery(acc, req.Systems, req.Advisories)
	var data []systemsAdvisoriesDBLoad
	if err := q.Find(&data).Error; err != nil {
		LogAndRespError(c, err, "Database error")
	}

	response := SystemsAdvisoriesResponse{
		Data: map[SystemID][]AdvisoryName{},
	}

	for _, i := range data {
		response.Data[i.SystemID] = append(response.Data[i.SystemID], i.AdvisoryID)
	}
	c.JSON(http.StatusOK, response)
}

// @Summary View advisory-system pairs for selected systems and advisories
// @Description View advisory-system pairs for selected systems and advisories
// @ID viewAdvisoriesSystems
// @Security RhIdentity
// @Accept   json
// @Produce  json
// @Param    body    body    SystemsAdvisoriesRequest true "Request body"
// @Success 200 {object} AdvisoriesSystemsResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /views/advisories/systems [post]
func PostAdvisoriesSystems(c *gin.Context) {
	var req SystemsAdvisoriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		LogAndRespBadRequest(c, err, "Invalid request body")
	}
	acc := c.GetInt(middlewares.KeyAccount)
	q := systemsAdvisoriesQuery(acc, req.Systems, req.Advisories)
	var data []systemsAdvisoriesDBLoad
	if err := q.Find(&data).Error; err != nil {
		LogAndRespError(c, err, "Database error")
	}

	response := AdvisoriesSystemsResponse{
		Data: map[AdvisoryName][]SystemID{},
	}

	for _, i := range data {
		response.Data[i.AdvisoryID] = append(response.Data[i.AdvisoryID], i.SystemID)
	}
	c.JSON(http.StatusOK, response)
}
