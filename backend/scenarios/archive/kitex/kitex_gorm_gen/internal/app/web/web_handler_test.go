package web

import (
	"context"
	"github.com/cloudwego/kitex-examples/bizdemo/kitex_gorm_gen/internal/repository/model/model"
	"testing"
)

var (
	testWeb = model.AnalyticsWeb{
		WebCode: "test_code",
		Name:    "test_name",
	}
	testAdv = model.AnalyticsWebAdvertiserRelation{
		AdvID: 9277,
	}
)

func init() {
	dao.dao.MustInitDB()
}

func TestCreateWeb(t *testing.T) {
	err := CreateWeb(context.Background(), testWeb.WebCode, testWeb.Name, int64(testAdv.AdvID))
	if err != nil {
		t.Errorf("CreateWeb failed: %v", err)
	}
	t.Logf("CreateWeb success")
}

func TestValidateWebParams(t *testing.T) {
	tests := []struct {
		advId   int64
		webCode string
		name    string
	}{
		{testAdv.AdvID, testWeb.WebCode, testWeb.Name},
		{9999, "non_exist_code", "non_exist_name"},
	}

	for _, test := range tests {
		web, err := ValidateWebParams(context.Background(), test.advId, test.webCode, test.name)
		if err != nil {
			t.Errorf("ValidateWebParams failed: %v", err)
		}
		t.Logf("ValidateWebParams success: %v", web)
	}
}

func TestUpdateWebAAMConfig(t *testing.T) {
	tests := []struct {
		advId   int64
		webCode string
		name    string
		extra   string
	}{
		{testAdv.AdvID, testWeb.WebCode, testWeb.Name, "test_extra"},
		{9999, "non_exist_code", "non_exist_name", "non_exist_extra"},
	}

	for _, test := range tests {
		err := UpdateWebAAMConfig(context.Background(), test.webCode, test.extra)
		if err != nil {
			t.Errorf("UpdateWebAAMConfig failed: %v", err)
		}
		t.Logf("UpdateWebAAMConfig finished")
	}
}

func TestDeleteWeb(t *testing.T) {
	tests := []struct {
		advId   int64
		webCode string
		name    string
		extra   string
	}{
		{testAdv.AdvID, testWeb.WebCode, testWeb.Name, "test_extra"},
		{9999, "non_exist_code", "non_exist_name", "non_exist_extra"},
	}

	for _, test := range tests {
		err := DeleteWeb(context.Background(), test.webCode)
		if err != nil {
			t.Errorf("DeleteWeb failed: %v", err)
		}
		t.Logf("DeleteWeb finished")
	}
}
