package controllers

import (
	"app/base/core"
	"app/base/utils"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdvisorySystemsDefault(t *testing.T) { //nolint:dupl
	core.SetupTest(t)
	w := CreateRequestRouterWithPath("GET", "/RH-1", nil, "", AdvisorySystemsListHandler, "/:advisory_id")

	var output AdvisorySystemsResponse
	CheckResponse(t, w, http.StatusOK, &output)
	assert.Equal(t, 6, len(output.Data))
	assert.Equal(t, "00000000-0000-0000-0000-000000000001", output.Data[0].ID)
	assert.Equal(t, "00000000-0000-0000-0001-000000000001", output.Data[0].Attributes.InsightsID)
	assert.Equal(t, "system", output.Data[0].Type)
	assert.Equal(t, "2018-09-22 16:00:00 +0000 UTC", output.Data[0].Attributes.LastEvaluation.String())
	assert.Equal(t, "2020-09-22 16:00:00 +0000 UTC", output.Data[0].Attributes.LastUpload.String())
	assert.False(t, output.Data[0].Attributes.Stale)
	assert.True(t, output.Data[0].Attributes.ThirdParty)
	assert.Equal(t, 2, output.Data[0].Attributes.RhsaCount)
	assert.Equal(t, 3, output.Data[0].Attributes.RhbaCount)
	assert.Equal(t, 3, output.Data[0].Attributes.RheaCount)
	assert.Equal(t, "RHEL", output.Data[0].Attributes.OSName)
	assert.Equal(t, "8", output.Data[0].Attributes.OSMajor)
	assert.Equal(t, "10", output.Data[0].Attributes.OSMinor)
	assert.Equal(t, "RHEL 8.10", output.Data[0].Attributes.OS)
	assert.Equal(t, "8.10", output.Data[0].Attributes.Rhsm)
	assert.Equal(t, "2018-08-26 16:00:00 +0000 UTC", output.Data[0].Attributes.StaleTimestamp.String())
	assert.Equal(t, "2018-09-02 16:00:00 +0000 UTC", output.Data[0].Attributes.StaleWarningTimestamp.String())
	assert.Equal(t, "2018-09-09 16:00:00 +0000 UTC", output.Data[0].Attributes.CulledTimestamp.String())
	assert.Equal(t, "2018-08-26 16:00:00 +0000 UTC", output.Data[0].Attributes.Created.String())
	assert.Equal(t, SystemTagsList{{"k1", "ns1", "val1"}, {"k2", "ns1", "val2"}}, output.Data[0].Attributes.Tags)
	assert.Equal(t, "baseline_1-1", output.Data[0].Attributes.BaselineName)
	assert.Equal(t, true, *output.Data[0].Attributes.BaselineUpToDate)
}

func TestAdvisorySystemsNotFound(t *testing.T) { //nolint:dupl
	core.SetupTest(t)
	w := CreateRequestRouterWithPath("GET", "/nonexistant/systems", nil, "", AdvisorySystemsListHandler, "/:advisory_id")

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAdvisorySystemsOffsetLimit(t *testing.T) { //nolint:dupl
	core.SetupTest(t)
	w := CreateRequestRouterWithPath("GET", "/RH-1?offset=5&limit=3", nil, "", AdvisorySystemsListHandler,
		"/:advisory_id")

	var output AdvisorySystemsResponse
	CheckResponse(t, w, http.StatusOK, &output)
	assert.Equal(t, 1, len(output.Data))
	assert.Equal(t, "00000000-0000-0000-0000-000000000006", output.Data[0].ID)
	assert.Equal(t, "system", output.Data[0].Type)
	assert.Equal(t, "2018-08-26 16:00:00 +0000 UTC", output.Data[0].Attributes.LastUpload.String())
	assert.Equal(t, 0, output.Data[0].Attributes.RhsaCount)
	assert.Equal(t, 1, output.Data[0].Attributes.RheaCount)
	assert.Equal(t, 0, output.Data[0].Attributes.RhbaCount)
}

func TestAdvisorySystemsOffsetOverflow(t *testing.T) {
	core.SetupTest(t)
	w := CreateRequestRouterWithPath("GET", "/RH-1?offset=100&limit=3", nil, "", AdvisorySystemsListHandler,
		"/:advisory_id")

	var errResp utils.ErrorResponse
	CheckResponse(t, w, http.StatusBadRequest, &errResp)
	assert.Equal(t, InvalidOffsetMsg, errResp.Error)
}

func TestAdvisorySystemsSorts(t *testing.T) {
	core.SetupTest(t)

	for sort := range SystemsFields {
		w := CreateRequestRouterWithPath("GET", fmt.Sprintf("/RH-1?sort=%v", sort), nil, "",
			AdvisorySystemsListHandler, "/:advisory_id")

		var output AdvisorySystemsResponse
		CheckResponse(t, w, http.StatusOK, &output)
		assert.Equal(t, 1, len(output.Meta.Sort))
		assert.Equal(t, output.Meta.Sort[0], sort)
	}
}

func TestAdvisorySystemsWrongSort(t *testing.T) { //nolint:dupl
	core.SetupTest(t)
	w := CreateRequestRouterWithPath("GET", "/RH-1?sort=unknown_key", nil, "", AdvisorySystemsListHandler,
		"/:advisory_id")

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdvisorySystemsTags(t *testing.T) { //nolint:dupl
	core.SetupTest(t)
	w := CreateRequestRouterWithPath("GET", "/RH-1?tags=ns1/k1=val1", nil, "", AdvisorySystemsListHandler,
		"/:advisory_id")

	var output AdvisorySystemsResponse
	CheckResponse(t, w, http.StatusOK, &output)
	assert.Equal(t, 5, len(output.Data))
}

func TestAdvisorySystemsTagsMultiple(t *testing.T) { //nolint:dupl
	core.SetupTest(t)
	w := CreateRequestRouterWithPath("GET", "/RH-1?tags=ns1/k3=val4&tags=ns1/k1=val1", nil, "",
		AdvisorySystemsListHandler, "/:advisory_id")

	var output AdvisorySystemsResponse
	CheckResponse(t, w, http.StatusOK, &output)
	assert.Equal(t, 1, len(output.Data))
	assert.Equal(t, "00000000-0000-0000-0000-000000000003", output.Data[0].ID)
}

func TestAdvisorySystemsTagsInvalid(t *testing.T) {
	core.SetupTest(t)
	w := CreateRequestRouterWithPath("GET", "/RH-1?tags=ns1/k3=val4&tags=invalidTag", nil, "",
		AdvisorySystemsListHandler, "/:advisory_id")

	var errResp utils.ErrorResponse
	CheckResponse(t, w, http.StatusBadRequest, &errResp)
	assert.Equal(t, fmt.Sprintf(InvalidTagMsg, "invalidTag"), errResp.Error)
}

func TestAdvisorySystemsTagsUnknown(t *testing.T) { //nolint:dupl
	core.SetupTest(t)
	w := CreateRequestRouterWithPath("GET", "/RH-1?tags=ns1/k3=val4&tags=ns1/k1=unk", nil, "",
		AdvisorySystemsListHandler, "/:advisory_id")

	var output AdvisorySystemsResponse
	CheckResponse(t, w, http.StatusOK, &output)
	assert.Equal(t, 0, len(output.Data))
}

func TestAdvisorySystemsWrongOffset(t *testing.T) {
	doTestWrongOffset(t, "/:advisory_id", "/RH-1?offset=1000", AdvisorySystemsListHandler)
}

func TestAdvisorySystemsTagsInMetadata(t *testing.T) {
	core.SetupTest(t)
	w := CreateRequestRouterWithPath("GET", "/RH-1?tags=ns1/k3=val4&tags=ns1/k1=val1", nil, "",
		AdvisorySystemsListHandler, "/:advisory_id")

	var output AdvisorySystemsResponse
	CheckResponse(t, w, http.StatusOK, &output)

	testMap := map[string]FilterData{
		"ns1/k1": {"eq", []string{"val1"}},
		"ns1/k3": {"eq", []string{"val4"}},
		"stale":  {"eq", []string{"false"}},
	}
	assert.Equal(t, testMap, output.Meta.Filter)
}
