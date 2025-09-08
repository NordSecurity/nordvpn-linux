package request

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

type PublishingRoundTripper struct {
	roundTripper http.RoundTripper
	publisher    events.Publisher[events.DataRequestAPI]
}

func NewPublishingRoundTripper(
	roundTripper http.RoundTripper,
	publisher events.Publisher[events.DataRequestAPI],
) *PublishingRoundTripper {
	return &PublishingRoundTripper{
		roundTripper: roundTripper,
		publisher:    publisher,
	}
}

type urlParameters struct {
	filters string
	fields  string
	limit   string
	offset  string
}

func getParameter(parameter string, values url.Values) string {
	parameterValue, ok := values[parameter]
	if !ok {
		return ""
	}

	if len(parameterValue) != 1 || parameterValue[0] == "" {
		log.Println(internal.WarningPrefix, "invalid value of limit parameter in the api call URL")
		return ""
	}

	return fmt.Sprintf("%s=%s", parameter, parameterValue[0])
}

func getURLParameters(url *url.URL) urlParameters {
	filters := []string{}
	fields := []string{}
	queryMap := url.Query()
	for query, value := range queryMap {
		isFilter := strings.Contains(query, "filters")
		isField := strings.Contains(query, "fields")
		if !isFilter && !isField {
			continue
		}

		var queryString string
		if len(value) != 0 && value[0] != "" {
			queryString = fmt.Sprintf("%s=%s", query, value[0])
		} else {
			queryString = query
		}

		if isFilter {
			filters = append(filters, queryString)
		}
		if isField {
			fields = append(fields, queryString)
		}
	}

	filtersString := strings.Join(filters, "&")
	fieldsString := strings.Join(fields, "&")

	limit := getParameter("limit", queryMap)
	offset := getParameter("offset", queryMap)

	return urlParameters{
		filters: filtersString,
		fields:  fieldsString,
		limit:   limit,
		offset:  offset,
	}
}

func (rt *PublishingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	parameters := getURLParameters(req.URL)

	startTime := time.Now()
	rt.publisher.Publish(events.DataRequestAPI{
		Request:   req,
		Error:     nil,
		Duration:  time.Since(startTime),
		IsAttempt: true,
	})
	resp, err := rt.roundTripper.RoundTrip(req)
	rt.publisher.Publish(events.DataRequestAPI{
		Request:        req,
		Response:       resp,
		Error:          err,
		Duration:       time.Since(startTime),
		IsAttempt:      false,
		RequestFilters: parameters.filters,
		RequestFields:  parameters.fields,
		Limits:         parameters.limit,
		Offset:         parameters.offset,
	})
	return resp, err
}
