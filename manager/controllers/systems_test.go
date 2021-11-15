package controllers

import (
	"app/base/core"
	"app/base/utils"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSystemsDefault(t *testing.T) {
	output := testSystems(t, `/`, 1)

	// data
	assert.Equal(t, 8, len(output.Data))
	assert.Equal(t, "00000000-0000-0000-0000-000000000001", output.Data[0].ID)
	assert.Equal(t, "00000000-0000-0000-0001-000000000001", output.Data[0].Attributes.InsightsID)
	assert.Equal(t, "system", output.Data[0].Type)
	assert.Equal(t, "2020-09-22 16:00:00 +0000 UTC", output.Data[0].Attributes.LastUpload.String())
	assert.Equal(t, "2018-09-22 16:00:00 +0000 UTC", output.Data[0].Attributes.LastEvaluation.String())
	assert.Equal(t, 3, output.Data[0].Attributes.RheaCount)
	assert.Equal(t, 3, output.Data[0].Attributes.RhbaCount)
	assert.Equal(t, 2, output.Data[0].Attributes.RhsaCount)
	assert.Equal(t, 0, output.Data[0].Attributes.PackagesInstalled)
	assert.Equal(t, 0, output.Data[0].Attributes.PackagesUpdatable)
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

	// links
	assert.Equal(t, "/api/patch/v1/systems?offset=0&limit=20&filter[stale]=eq:false&sort=-last_upload", output.Links.First)
	assert.Equal(t, "/api/patch/v1/systems?offset=0&limit=20&filter[stale]=eq:false&sort=-last_upload", output.Links.Last)
	assert.Nil(t, output.Links.Next)
	assert.Nil(t, output.Links.Previous)

	// test meta
	assert.Equal(t, 0, output.Meta.Offset)
	assert.Equal(t, core.DefaultLimit, output.Meta.Limit)
	assert.Equal(t, 8, output.Meta.TotalItems)
	assert.Equal(t, 8, output.Meta.SubTotals["patched"])
	assert.Equal(t, 0, output.Meta.SubTotals["unpatched"])
	assert.Equal(t, 0, output.Meta.SubTotals["stale"])
}

func TestSystemsOffset(t *testing.T) {
	output := testSystems(t, `/?offset=0&limit=4`, 1)
	assert.Equal(t, 4, len(output.Data))
	assert.Equal(t, 0, output.Meta.Offset)
	assert.Equal(t, 4, output.Meta.Limit)
	assert.Equal(t, 8, output.Meta.TotalItems)
}

func TestSystemsOffsetLimit(t *testing.T) {
	output := testSystems(t, `/?offset=4&limit=4`, 1)
	assert.Equal(t, 4, len(output.Data))
	assert.Equal(t, 4, output.Meta.Offset)
	assert.Equal(t, 4, output.Meta.Limit)
	assert.Equal(t, 8, output.Meta.TotalItems)
}

func TestSystemsWrongOffset(t *testing.T) {
	doTestWrongOffset(t, "/", "/?offset=13&limit=4", SystemsListHandler)
}

func TestSystemsWrongSort(t *testing.T) {
	statusCode, errResp := testSystemsError(t, "/?sort=unknown_key")
	assert.Equal(t, http.StatusBadRequest, statusCode)
	assert.Equal(t, "Invalid sort field: unknown_key", errResp.Error)
}

func TestSystemsSearch(t *testing.T) {
	output := testSystems(t, "/?search=001", 1)
	assert.Equal(t, 1, len(output.Data))
	assert.Equal(t, "00000000-0000-0000-0000-000000000001", output.Data[0].ID)
	assert.Equal(t, "system", output.Data[0].Type)
	assert.Equal(t, "00000000-0000-0000-0000-000000000001", output.Data[0].Attributes.DisplayName)
}

func TestSystemsTags(t *testing.T) {
	output := testSystems(t, "/?tags=ns1/k2=val2", 1)
	assert.Equal(t, 2, len(output.Data))
	assert.Equal(t, "00000000-0000-0000-0000-000000000001", output.Data[0].ID)
}

func TestSystemsTagsMultiple(t *testing.T) {
	output := testSystems(t, "/?tags=ns1/k3=val4&tags=ns1/k1=val1", 1)
	assert.Equal(t, 1, len(output.Data))
	assert.Equal(t, "00000000-0000-0000-0000-000000000003", output.Data[0].ID)
}

func TestSystemsTagsUnknown(t *testing.T) {
	output := testSystems(t, "/?tags=ns1/k3=val4&tags=ns1/k1=unk", 1)
	assert.Equal(t, 0, len(output.Data))
}

func TestSystemsTagsNoVal(t *testing.T) {
	output := testSystems(t, "/?tags=ns1/k3=val4&tags=ns1/k1", 1)
	assert.Equal(t, 1, len(output.Data))
	assert.Equal(t, "00000000-0000-0000-0000-000000000003", output.Data[0].ID)
}

func TestSystemsTagsInvalid(t *testing.T) {
	statusCode, errResp := testSystemsError(t, "/?tags=ns1/k3=val4&tags=invalidTag")
	assert.Equal(t, http.StatusBadRequest, statusCode)
	assert.Equal(t, fmt.Sprintf(InvalidTagMsg, "invalidTag"), errResp.Error)
}

func TestSystemsWorkloads1(t *testing.T) {
	url := "/?filter[system_profile][sap_system]=true&filter[system_profile][sap_sids][in][]=ABC"
	output := testSystems(t, url, 1)
	assert.Equal(t, 2, len(output.Data))
	assert.Equal(t, "00000000-0000-0000-0000-000000000001", output.Data[0].ID)
}

func TestSystemsWorkloads2(t *testing.T) {
	url := "/?filter[system_profile][sap_system]=true&filter[system_profile][sap_sids][]=ABC"
	output := testSystems(t, url, 1)
	assert.Equal(t, 2, len(output.Data))
	assert.Equal(t, "00000000-0000-0000-0000-000000000001", output.Data[0].ID)
}

func TestSystemsWorkloads3(t *testing.T) {
	output := testSystems(t, "/?filter[system_profile][sap_system]=false", 1)
	assert.Equal(t, 0, len(output.Data))
}

func TestSystemsPackagesCount(t *testing.T) {
	output := testSystems(t, "/?sort=-packages_installed,id", 3)
	assert.Equal(t, 3, len(output.Data))
	assert.Equal(t, "00000000-0000-0000-0000-000000000012", output.Data[0].ID)
	assert.Equal(t, "system", output.Data[0].Type)
	assert.Equal(t, "00000000-0000-0000-0000-000000000012", output.Data[0].Attributes.DisplayName)
	assert.Equal(t, 2, output.Data[0].Attributes.PackagesInstalled)
	assert.Equal(t, 2, output.Data[0].Attributes.PackagesUpdatable)
}

func TestSystemsFilterAdvCount1(t *testing.T) {
	output := testSystems(t, "/?filter[rhba_count]=3", 1)
	assert.Equal(t, 1, len(output.Data))
	assert.Equal(t, 3, output.Data[0].Attributes.RhbaCount)
}

func TestSystemsFilterAdvCount2(t *testing.T) {
	output := testSystems(t, "/?filter[rhea_count]=3", 1)
	assert.Equal(t, 1, len(output.Data))
	assert.Equal(t, 3, output.Data[0].Attributes.RheaCount)
}

func TestSystemsFilterAdvCount3(t *testing.T) {
	output := testSystems(t, "/?filter[rhsa_count]=2", 1)
	assert.Equal(t, 1, len(output.Data))
	assert.Equal(t, 2, output.Data[0].Attributes.RhsaCount)
}

func TestSystemsFilterAdvCount4(t *testing.T) {
	output := testSystems(t, "/?filter[other_count]=1", 1)
	assert.Equal(t, 1, len(output.Data))
	assert.Equal(t, 1, output.Data[0].Attributes.OtherCount)
}

func TestSystemsFilterNotExisting(t *testing.T) {
	statusCode, errResp := testSystemsError(t, "/?filter[not-existing]=1")
	assert.Equal(t, http.StatusBadRequest, statusCode)
	assert.Equal(t, "Invalid filter field: not-existing", errResp.Error)
}

func TestSystemsFilterPartialOS(t *testing.T) {
	output := testSystems(t, "/?filter[osname]=RHEL&filter[osmajor]=8&filter[osminor]=1", 1)
	assert.Equal(t, 2, len(output.Data))
	assert.Equal(t, "RHEL 8.1", fmt.Sprintf("%s %s.%s", output.Data[0].Attributes.OSName,
		output.Data[0].Attributes.OSMajor, output.Data[0].Attributes.OSMinor))
	assert.Equal(t, "RHEL 8.1", fmt.Sprintf("%s %s.%s", output.Data[1].Attributes.OSName,
		output.Data[1].Attributes.OSMajor, output.Data[1].Attributes.OSMinor))
}

func TestSystemsFilterOS(t *testing.T) {
	output := testSystems(t, `/?filter[os]=in:RHEL 8.1,RHEL 7.3&sort=os`, 1)
	assert.Equal(t, 3, len(output.Data))
	assert.Equal(t, "RHEL 7.3", output.Data[0].Attributes.OS)
	assert.Equal(t, "RHEL 8.1", output.Data[1].Attributes.OS)
	assert.Equal(t, "RHEL 8.1", output.Data[2].Attributes.OS)
}

func TestSystemsFilterInvalidSyntax(t *testing.T) {
	statusCode, errResp := testSystemsError(t, "/?filter[os][in]=RHEL 8.1,RHEL 7.3")
	assert.Equal(t, http.StatusBadRequest, statusCode)
	assert.Equal(t, InvalidNestedFilter, errResp.Error)
}

func TestSystemsOrderOS(t *testing.T) {
	output := testSystems(t, `/?sort=os`, 1)
	assert.Equal(t, "RHEL 7.3", output.Data[0].Attributes.OS)
	assert.Equal(t, "RHEL 8.x", output.Data[1].Attributes.OS) // yes, we should be robust against this
	assert.Equal(t, "RHEL 8.1", output.Data[2].Attributes.OS)
	assert.Equal(t, "RHEL 8.1", output.Data[3].Attributes.OS)
	assert.Equal(t, "RHEL 8.2", output.Data[4].Attributes.OS)
	assert.Equal(t, "RHEL 8.3", output.Data[5].Attributes.OS)
	assert.Equal(t, "RHEL 8.3", output.Data[6].Attributes.OS)
	assert.Equal(t, "RHEL 8.10", output.Data[7].Attributes.OS)
}

func testSystems(t *testing.T, url string, account int) SystemsResponse {
	utils.SkipWithoutDB(t)
	core.SetupTestEnvironment()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", url, nil)
	core.InitRouterWithAccount(SystemsListHandler, "/", account).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var output SystemsResponse
	ParseReponseBody(t, w.Body.Bytes(), &output)
	return output
}

func testSystemsError(t *testing.T, url string) (int, utils.ErrorResponse) {
	utils.SkipWithoutDB(t)
	core.SetupTestEnvironment()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", url, nil)
	core.InitRouterWithPath(SystemsListHandler, "/").ServeHTTP(w, req)
	var errResp utils.ErrorResponse
	ParseReponseBody(t, w.Body.Bytes(), &errResp)
	return w.Code, errResp
}
