package svcconfig

func init() {
	SchemaRegistry["ysfgateway"] = &SettingsSchema{
		ServiceName: "ysfgateway",
		Groups: []SettingsGroup{
			{
				Name:    "General",
				I18nKey: "config.ysfgateway.general",
				Fields: []SettingsField{
					{
						Key:        "suffix",
						I18nLabel:  "config.ysfgateway.suffix",
						INISection: "General",
						INIKey:     "Suffix",
						FieldType:  "text",
						Validate:   "maxlen:7",
						Default:    "RPT",
					},
				},
			},
			{
				Name:    "Network",
				I18nKey: "config.ysfgateway.network",
				Fields: []SettingsField{
					{
						Key:        "startup",
						I18nLabel:  "config.ysfgateway.startup",
						INISection: "Network",
						INIKey:     "Startup",
						FieldType:  "text",
						Default:    "",
						HelpI18n:   "config.ysfgateway.startup.help",
					},
					{
						Key:        "inactivityTimeout",
						I18nLabel:  "config.ysfgateway.inactivityTimeout",
						INISection: "Network",
						INIKey:     "InactivityTimeout",
						FieldType:  "number",
						Validate:   "numeric,range:0:600",
						Default:    "10",
						HelpI18n:   "config.ysfgateway.inactivityTimeout.help",
					},
				},
			},
		},
	}
}
