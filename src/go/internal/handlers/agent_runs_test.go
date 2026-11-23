package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestGoalFilterFromQuery kontrollerar att query-parametrarna blir rätt
// filter - och att en trasig parameter ger ett fel.
func TestGoalFilterFromQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name       string
		query      string
		wantAgent  *bool
		wantStatus *int
		wantErr    bool
	}{
		{name: "inga filter", query: ""},
		{name: "agent true", query: "?agent=true", wantAgent: boolPtr(true)},
		{name: "agent false", query: "?agent=false", wantAgent: boolPtr(false)},
		{name: "status", query: "?status=3", wantStatus: intPtr(3)},
		{name: "båda", query: "?agent=true&status=2", wantAgent: boolPtr(true), wantStatus: intPtr(2)},
		{name: "trasig agent", query: "?agent=ja", wantErr: true},
		{name: "trasig status", query: "?status=abc", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest("GET", "/api/v1/goals"+tc.query, nil)

			filter, err := goalFilterFromQuery(c)
			if tc.wantErr {
				if err == nil {
					t.Fatal("väntade ett fel")
				}
				return
			}
			if err != nil {
				t.Fatalf("oväntat fel: %v", err)
			}

			if tc.wantAgent == nil {
				if filter.Agent != nil {
					t.Errorf("agent-filter ska vara nil, fick %v", *filter.Agent)
				}
			} else if filter.Agent == nil || *filter.Agent != *tc.wantAgent {
				t.Errorf("agent-filter = %v, vill ha %v", filter.Agent, *tc.wantAgent)
			}

			if tc.wantStatus == nil {
				if filter.StatusID != nil {
					t.Errorf("status-filter ska vara nil, fick %v", *filter.StatusID)
				}
			} else if filter.StatusID == nil || *filter.StatusID != *tc.wantStatus {
				t.Errorf("status-filter = %v, vill ha %v", filter.StatusID, *tc.wantStatus)
			}
		})
	}
}

func boolPtr(v bool) *bool { return &v }
func intPtr(v int) *int    { return &v }
