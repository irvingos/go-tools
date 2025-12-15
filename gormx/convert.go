package gormx

import (
	"encoding/json"

	"gorm.io/datatypes"
)

func JSONToMap(j datatypes.JSON) (map[string]any, error) {
	if len(j) == 0 {
		return nil, nil
	}

	var m map[string]any
	if err := json.Unmarshal(j, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func MapToJSON(m map[string]any) (datatypes.JSON, error) {
	if m == nil {
		return nil, nil
	}

	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(b), nil
}

func JSONToSliceMap(j datatypes.JSON) ([]map[string]any, error) {
	if len(j) == 0 {
		return nil, nil
	}

	var s []map[string]any
	if err := json.Unmarshal(j, &s); err != nil {
		return nil, err
	}
	return s, nil
}

func SliceMapToJSON(s []map[string]any) (datatypes.JSON, error) {
	if s == nil {
		return nil, nil
	}

	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(b), nil
}

func JSONTo[T any](j datatypes.JSON, v *T) error {
	if len(j) == 0 {
		return nil
	}
	return json.Unmarshal(j, v)
}

func ToJSON[T any](v T) (datatypes.JSON, error) {
	b, err := json.Marshal(v)
	return datatypes.JSON(b), err
}

func JSONToAny(j datatypes.JSON) (any, error) {
	if len(j) == 0 {
		return nil, nil
	}

	var v any
	err := json.Unmarshal(j, &v)
	return v, err
}

func AnyToJSON(v any) (datatypes.JSON, error) {
	if v == nil {
		return nil, nil
	}

	b, err := json.Marshal(v)
	return datatypes.JSON(b), err
}
