package svcconfig

func init() {
	SchemaRegistry["mmdvmhost"] = &SettingsSchema{
		ServiceName: "mmdvmhost",
		Groups: []SettingsGroup{
			{
				Name:    "DMR",
				I18nKey: "config.mmdvmhost.dmr",
				Fields: []SettingsField{
					{
						Key:        "colorCode",
						I18nLabel:  "config.mmdvmhost.colorCode",
						INISection: "DMR",
						INIKey:     "ColorCode",
						FieldType:  "number",
						Validate:   "numeric,range:1:15",
						Default:    "1",
					},
				},
			},
			{
				Name:    "D-Star",
				I18nKey: "config.mmdvmhost.dstar",
				Fields: []SettingsField{
					{
						Key:        "dstarModule",
						I18nLabel:  "config.mmdvmhost.dstarModule",
						INISection: "D-Star",
						INIKey:     "Module",
						FieldType:  "select",
						Default:    "C",
						Options: []Option{
							{Value: "A", I18nKey: "config.mmdvmhost.dstarModule.a"},
							{Value: "B", I18nKey: "config.mmdvmhost.dstarModule.b"},
							{Value: "C", I18nKey: "config.mmdvmhost.dstarModule.c"},
						},
					},
				},
			},
		},
	}
}
