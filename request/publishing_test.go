package request

import (
	"net/url"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

const testEndpoint string = "https://test.api.com/endpoint"

func getURL(parameters ...string) string {
	url := testEndpoint + "/?"
	parametersString := strings.Join(parameters, "&")

	return url + parametersString
}

func areParametersEqual(a string, b string) bool {
	if len(a) != len(b) {
		return false
	}

	for _, parameter := range strings.Split(a, "&") {
		if !strings.Contains(b, parameter) {
			return false
		}
	}

	return true
}

func TestGetQueryParameters(t *testing.T) {
	filter1 := "filters[filter1]=value"
	field1 := "fields[field1]"
	filter2 := "filters[filter2]=value"
	field2 := "fields[field2]=12"

	offset := "offset=20"
	invalidOffset := "offset"

	limit := "limit=10"
	invalidLimit := "limit"

	tests := []struct {
		name               string
		url                string
		expectedParameters urlParameters
	}{
		{
			name: "no parameters",
			url:  testEndpoint,
		},
		{
			name: "only offset parameter",
			url:  getURL(offset),
			expectedParameters: urlParameters{
				offset: offset,
			},
		},
		{
			name: "offset and limit parameters",
			url:  getURL(offset, limit),
			expectedParameters: urlParameters{
				offset: offset,
				limit:  limit,
			},
		},
		{
			name: "one filter and one field parameter",
			url:  getURL(offset, limit, filter1, field1),
			expectedParameters: urlParameters{
				filters: filter1,
				fields:  field1,
				offset:  offset,
				limit:   limit,
			},
		},
		{
			name: "two filters and two field parameter",
			url:  getURL(offset, limit, filter1, filter2, field1, field2),
			expectedParameters: urlParameters{
				filters: filter1 + "&" + filter2,
				fields:  field1 + "&" + field2,
				offset:  offset,
				limit:   limit,
			},
		},
		{
			name: "filter and field parameters are not grouped together",
			url:  getURL(offset, limit, filter1, field1, filter2, field2),
			expectedParameters: urlParameters{
				filters: filter1 + "&" + filter2,
				fields:  field1 + "&" + field2,
				offset:  offset,
				limit:   limit,
			},
		},
		{
			name: "invalid offset and limit is ignored",
			url:  getURL(invalidOffset, invalidLimit),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			url, _ := url.Parse(test.url)
			parameters := getURLParameters(url)
			assert.Equal(t,
				true,
				areParametersEqual(test.expectedParameters.filters, parameters.filters),
				"Generated parameters lists are not equal: %s != %s", test.expectedParameters.filters, parameters)
			assert.Equal(t,
				true,
				areParametersEqual(test.expectedParameters.fields, parameters.fields),
				"Generated parameters lists are not equal: %s != %s", test.expectedParameters.fields, parameters)
			assert.Equal(t, test.expectedParameters.offset, parameters.offset)
			assert.Equal(t, test.expectedParameters.limit, parameters.limit)
		})
	}
}
