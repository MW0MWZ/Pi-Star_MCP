package svcconfig

import "testing"

func TestAllSchemasHaveValidFields(t *testing.T) {
	if len(SchemaRegistry) == 0 {
		t.Fatal("SchemaRegistry is empty — no schemas registered")
	}

	for name, schema := range SchemaRegistry {
		if schema.ServiceName == "" {
			t.Errorf("schema %q has empty ServiceName", name)
		}
		if len(schema.Groups) == 0 {
			t.Errorf("schema %q has no groups", name)
		}

		keys := make(map[string]bool)
		for _, g := range schema.Groups {
			if g.Name == "" {
				t.Errorf("schema %q has a group with empty Name", name)
			}
			if g.I18nKey == "" {
				t.Errorf("schema %q group %q has empty I18nKey", name, g.Name)
			}
			if len(g.Fields) == 0 {
				t.Errorf("schema %q group %q has no fields", name, g.Name)
			}

			for _, f := range g.Fields {
				if f.Key == "" {
					t.Errorf("schema %q group %q has a field with empty Key", name, g.Name)
				}
				if keys[f.Key] {
					t.Errorf("schema %q has duplicate field key %q", name, f.Key)
				}
				keys[f.Key] = true

				if f.INISection == "" {
					t.Errorf("schema %q field %q has empty INISection", name, f.Key)
				}
				if f.INIKey == "" {
					t.Errorf("schema %q field %q has empty INIKey", name, f.Key)
				}
				if f.I18nLabel == "" {
					t.Errorf("schema %q field %q has empty I18nLabel", name, f.Key)
				}

				validTypes := map[string]bool{"text": true, "number": true, "boolean": true, "select": true}
				if !validTypes[f.FieldType] {
					t.Errorf("schema %q field %q has invalid FieldType %q", name, f.Key, f.FieldType)
				}

				if f.FieldType == "select" && len(f.Options) == 0 {
					t.Errorf("schema %q field %q is select type but has no Options", name, f.Key)
				}
			}
		}
	}
}

func TestLookupSchema(t *testing.T) {
	s, ok := LookupSchema("mmdvmhost")
	if !ok {
		t.Fatal("expected mmdvmhost schema to exist")
	}
	if s.ServiceName != "mmdvmhost" {
		t.Errorf("expected ServiceName mmdvmhost, got %q", s.ServiceName)
	}

	_, ok = LookupSchema("nonexistent")
	if ok {
		t.Error("expected nonexistent schema lookup to return false")
	}
}

func TestRegisteredSchemas(t *testing.T) {
	expected := []string{"mmdvmhost", "dmrgateway", "ysfgateway"}
	for _, name := range expected {
		if _, ok := SchemaRegistry[name]; !ok {
			t.Errorf("expected schema %q to be registered", name)
		}
	}
}
