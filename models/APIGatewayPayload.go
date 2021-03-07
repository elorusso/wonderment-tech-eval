package models

type APIGatewayPayload struct {
	Version               string
	RouteKey              string
	RawPath               string
	RawQueryString        string
	Cookies               []string
	Headers               map[string]string
	QueryStringParameters map[string]string
	StageVariables        map[string]string
	Body                  string
	PathParameters        map[string]string
	RequestContext        *RequestContext
}

type RequestContext struct {
	AccountID    string
	APIID        string
	DomainName   string
	DomainPrefix string
	HTTP         *HTTPInfo
	RequestID    string
	RouteKey     string
	Stage        string
	TimeEpoch    int
}

type HTTPInfo struct {
	Method    string
	Path      string
	Protocol  string
	SourceIP  string
	UserAgent string
}
