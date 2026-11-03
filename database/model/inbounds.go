package model

import (
	"encoding/json"
	"fmt"
)

type Inbound struct {
	Id   uint   `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	Type string `json:"type" form:"type"`
	Tag  string `json:"tag" form:"tag" gorm:"unique"`

	// Foreign key to tls table
	TlsId uint `json:"tls_id" form:"tls_id"`
	Tls   *Tls `json:"tls" form:"tls" gorm:"foreignKey:TlsId;references:Id"`

	Addrs   json.RawMessage `json:"addrs" form:"addrs"`
	OutJson json.RawMessage `json:"out_json" form:"out_json"`
	Options json.RawMessage `json:"-" form:"-"`
}

func (i *Inbound) UnmarshalJSON(data []byte) error {
	var err error
	var raw map[string]interface{}
	if err = json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Extract fixed fields and store the rest in Options
	if val, exists := raw["id"].(float64); exists {
		i.Id = uint(val)
	}
	delete(raw, "id")
	i.Type, _ = raw["type"].(string)
	delete(raw, "type")
	i.Tag, _ = raw["tag"].(string)
	delete(raw, "tag")

	// TlsId
	if val, exists := raw["tls_id"].(float64); exists {
		i.TlsId = uint(val)
	}
	delete(raw, "tls_id")
	delete(raw, "tls")
	delete(raw, "users")

	// Addrs
	i.Addrs, _ = json.MarshalIndent(raw["addrs"], "", "  ")
	delete(raw, "addrs")

	// OutJson
	i.OutJson, _ = json.MarshalIndent(raw["out_json"], "", "  ")
	delete(raw, "out_json")

	// Remaining fields
	i.Options, err = json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	_, err = i.IsEnabled()
	return err
}

// Enable is panel metadata stored in Options. Existing inbounds without the
// field (or with null) remain enabled; sing-box never receives this field.
func (i Inbound) IsEnabled() (bool, error) {
	if len(i.Options) == 0 {
		return true, nil
	}
	var fields struct {
		Enable *bool `json:"enable"`
	}
	if err := json.Unmarshal(i.Options, &fields); err != nil {
		return false, fmt.Errorf("inbound %q enable: %w", i.Tag, err)
	}
	return fields.Enable == nil || *fields.Enable, nil
}

func (i *Inbound) SetEnabled(enabled bool) error {
	fields := make(map[string]json.RawMessage)
	if len(i.Options) > 0 {
		if err := json.Unmarshal(i.Options, &fields); err != nil {
			return err
		}
	}
	if fields == nil {
		fields = make(map[string]json.RawMessage)
	}
	fields["enable"], _ = json.Marshal(enabled)
	var err error
	i.Options, err = json.MarshalIndent(fields, "", "  ")
	return err
}

// MarshalJSON customizes marshalling
func (i Inbound) MarshalJSON() ([]byte, error) {
	// Combine fixed fields and dynamic fields into one map
	combined := make(map[string]interface{})
	combined["type"] = i.Type
	combined["tag"] = i.Tag
	if i.Tls != nil {
		combined["tls"] = i.Tls.Server
	}

	if len(i.Options) > 0 {
		var restFields map[string]json.RawMessage
		if err := json.Unmarshal(i.Options, &restFields); err != nil {
			return nil, err
		}

		for k, v := range restFields {
			combined[k] = v
		}
	}

	delete(combined, "enable")
	return json.Marshal(combined)
}

func (i Inbound) MarshalFull() (*map[string]interface{}, error) {
	combined := make(map[string]interface{})
	combined["id"] = i.Id
	combined["type"] = i.Type
	combined["tag"] = i.Tag
	combined["tls_id"] = i.TlsId
	combined["addrs"] = i.Addrs
	combined["out_json"] = i.OutJson

	if len(i.Options) > 0 {
		var restFields map[string]interface{}
		if err := json.Unmarshal(i.Options, &restFields); err != nil {
			return nil, err
		}

		for k, v := range restFields {
			combined[k] = v
		}
	}
	enabled, err := i.IsEnabled()
	if err != nil {
		return nil, err
	}
	combined["enable"] = enabled
	return &combined, nil
}
