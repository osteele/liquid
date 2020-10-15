package customtags

import "fmt"

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
	DisplayType  string  `json:"displayType" bson:"displayType"`
	LiquidName   string  `json:"liquidName" bson:"liquidName"`
	DefaultValue string  `json:"defaultValue" bson:"defaultValue"`
	FormatOption string  `json:"formatOption" bson:"formatOption"`
	FieldID      string  `json:"namespaceId" bson:"namespaceId"`
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
		case "date":
			return fmt.Sprintf(`{{ %s | dateFormatOrDefault: "%s", "%s" }}`, t.LiquidName, t.FormatOption, t.DefaultValue), nil
		case "time":
			return fmt.Sprintf(`{{ %s | dateTimeFormatOrDefault: "%s", "%s" }}`, t.LiquidName, t.FormatOption, t.DefaultValue), nil
		case "bool":
			return fmt.Sprintf(`{{ %s | booleanFormat: "%s" }}`, t.LiquidName, t.FormatOption), nil
		case "currency", "decimal", "aggregate":
			return fmt.Sprintf(`{{ %s | decimal: "%s", "%s" }}`, t.LiquidName, t.FormatOption, t.DefaultValue), nil
		case "phone":
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
