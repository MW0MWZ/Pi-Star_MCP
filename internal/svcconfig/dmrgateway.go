package svcconfig

func init() {
	SchemaRegistry["dmrgateway"] = &SettingsSchema{
		ServiceName: "dmrgateway",
		Groups: []SettingsGroup{
			{
				Name:    "General",
				I18nKey: "config.dmrgateway.general",
				Fields: []SettingsField{
					{
						Key:        "rptAddress",
						I18nLabel:  "config.dmrgateway.rptAddress",
						INISection: "General",
						INIKey:     "RptAddress",
						FieldType:  "text",
						Validate:   "required,ip",
						Default:    "127.0.0.1",
					},
					{
						Key:        "rptPort",
						I18nLabel:  "config.dmrgateway.rptPort",
						INISection: "General",
						INIKey:     "RptPort",
						FieldType:  "number",
						Validate:   "required,port",
						Default:    "62032",
					},
				},
			},
			{
				Name:    "DMR Network 1",
				I18nKey: "config.dmrgateway.network1",
				Fields: []SettingsField{
					{
						Key:        "net1Enabled",
						I18nLabel:  "config.dmrgateway.net1Enabled",
						INISection: "DMR Network 1",
						INIKey:     "Enabled",
						FieldType:  "boolean",
						Default:    "1",
					},
					{
						Key:        "net1Name",
						I18nLabel:  "config.dmrgateway.net1Name",
						INISection: "DMR Network 1",
						INIKey:     "Name",
						FieldType:  "text",
						Validate:   "required,maxlen:40",
						Default:    "BrandMeister",
					},
					{
						Key:        "net1Address",
						I18nLabel:  "config.dmrgateway.net1Address",
						INISection: "DMR Network 1",
						INIKey:     "Address",
						FieldType:  "text",
						Validate:   "required,hostname",
						Default:    "127.0.0.1",
					},
					{
						Key:        "net1Port",
						I18nLabel:  "config.dmrgateway.net1Port",
						INISection: "DMR Network 1",
						INIKey:     "Port",
						FieldType:  "number",
						Validate:   "required,port",
						Default:    "62031",
					},
					{
						Key:        "net1Password",
						I18nLabel:  "config.dmrgateway.net1Password",
						INISection: "DMR Network 1",
						INIKey:     "Password",
						FieldType:  "text",
						Validate:   "required",
						Default:    "passw0rd",
					},
				},
			},
		},
	}
}
