package customtags

import "fmt"

const (
	TagDisplayTypeDate      string = "date"
	TagDisplayTypeTime      string = "time"
	TagDisplayTypeBool      string = "bool"
	TagDisplayTypeCurrency  string = "currency"
	TagDisplayTypeDecimal   string = "decimal"
	TagDisplayTypeAggregate string = "aggregate"
	TagDisplayTypePhone     string = "phone"
)

type TagType string

const (
	TagTypeMergeTag  TagType = "merge"
	TagTypeMergeText TagType = "text"
)

type Tag struct {
	ID           string  `json:"id" bson:"id"`
	Type         TagType `json:"type" bson:"type"`
	Icon         string  `json:"icon" bson:"icon"`
	Title        string  `json:"title" bson:"title"`
	DisplayType  string  `json:"display_type" bson:"display_type"`
	LiquidName   string  `json:"liquid_name" bson:"liquid_name"`
	DefaultValue string  `json:"default_value" bson:"default_value"`
	FormatOption string  `json:"format_option" bson:"format_option"`
	FieldID      string  `json:"field_id" bson:"field_id"`
}

func (t *Tag) GetPreviewString() (string, error) {
	if t.Type == TagTypeMergeText {
		return t.Title, nil
	}
	if t.Type == TagTypeMergeTag {
		return fmt.Sprintf("[%s]", t.Title), nil
	}
	return "", fmt.Errorf("unsupported subject merge tag %+v type: %s", t, t.Type)
}

func (t *Tag) GetLiquidString() (string, error) {
	if t.Type == TagTypeMergeText {
		return t.Title, nil
	}
	if t.Type == TagTypeMergeTag {
		switch t.DisplayType {
		case TagDisplayTypeDate:
			return fmt.Sprintf(`{{ %s | dateFormatOrDefault: "%s", "%s" }}`, t.LiquidName, t.FormatOption, t.DefaultValue), nil
		case TagDisplayTypeTime:
			return fmt.Sprintf(`{{ %s | dateTimeFormatOrDefault: "%s", "%s" }}`, t.LiquidName, t.FormatOption, t.DefaultValue), nil
		case TagDisplayTypeBool:
			return fmt.Sprintf(`{{ %s | booleanFormat: "%s" }}`, t.LiquidName, t.FormatOption), nil
		case TagDisplayTypeCurrency, TagDisplayTypeDecimal, TagDisplayTypeAggregate:
			return fmt.Sprintf(`{{ %s | decimal: "%s", "%s" }}`, t.LiquidName, t.FormatOption, t.DefaultValue), nil
		case TagDisplayTypePhone:
			willHide := false
			if t.FormatOption == "hide" {
				willHide = true
			}
			return fmt.Sprintf(`{{ %s | hideCountryCodeAndDefault: %t, "%s" }}`, t.LiquidName, willHide, t.DefaultValue), nil
		default:
			result := `{{ ` + t.LiquidName
			if t.DefaultValue != "" {
				result += ` | default: "` + t.DefaultValue + `"`
			}
			result += ` }}`
			return result, nil
		}
	}
	return "", fmt.Errorf("unsupported subject merge tag %+v type: %s", t, t.Type)
}
