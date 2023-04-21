package response

import "github.com/tensuqiuwulu/be-service-teman-bunda/models/entity"

type FindSettingVerApp struct {
	SettingName string `json:"os_type"`
	Value       string `json:"ver_app"`
}

type FindSettingVerApp2 struct {
	OS      string `json:"os"`
	Current string `json:"current"`
	New     string `json:"new"`
}

func ToFindSettingVerAppList2(verApp []entity.Settings, os int) (verAppResponse FindSettingVerApp2) {
	if os == 1 {
		verAppResponse.OS = "android"
	} else {
		verAppResponse.OS = "ios"
	}

	verAppResponse.Current = verApp[0].SettingsTitle
	verAppResponse.New = verApp[0].SettingsTitle
	return verAppResponse
}

func ToFindSettingVerApp(verApp entity.Settings) (verAppResponse FindSettingVerApp) {
	verAppResponse.SettingName = verApp.SettingsName
	verAppResponse.Value = verApp.SettingsTitle
	return verAppResponse
}

func ToFindSettingVerAppList(verApp []entity.Settings) (verAppResponse []FindSettingVerApp) {
	for _, value := range verApp {
		var findSettingVerApp FindSettingVerApp
		findSettingVerApp.SettingName = value.SettingsName
		findSettingVerApp.Value = value.SettingsTitle
		verAppResponse = append(verAppResponse, ToFindSettingVerApp(value))
	}
	return verAppResponse
}
