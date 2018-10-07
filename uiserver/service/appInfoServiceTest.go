package service

import (
	"fmt"
	"testing"

	"github.com/a526757124/polaris/uiserver/viewModel"
)

var appInfoService *AppInfoService

func init() {
	appInfoService = &AppInfoService{}
}
func TestSum(t *testing.T) {
	query := new(viewModel.AppInfoQueryParm)
	query.PageIndex = 1
	query.PageSize = 10
	list, _ := new(AppInfoService).GetList(query)
	fmt.Println(list)
}
