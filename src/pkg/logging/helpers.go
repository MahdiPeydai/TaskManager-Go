package logging

func MapToZapParams(extras map[ExtraKey]interface{}) []interface{} {
	params := make([]interface{}, 0)
	for key, value := range extras {
		params = append(params, string(key))
		params = append(params, value)
	}
	return params
}

func MapToZeroParams(extras map[ExtraKey]interface{}) map[string]interface{} {
	params := make(map[string]interface{})
	for key, value := range extras {
		params[string(key)] = value
	}
	return params
}
