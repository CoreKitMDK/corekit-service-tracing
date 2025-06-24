package tracing

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Trace struct {
	UID               string
	RequestTimestamp  time.Time
	ResponseTimestamp time.Time
	TraceTimestamp    time.Time
	Request           string
	Response          string
	ServicePath       string
	TotalRequestTime  string
	Hop               int
	Tags              map[string]string
}

func NewTrace() *Trace {
	return &Trace{
		UID:               "",
		RequestTimestamp:  time.Now(),
		ResponseTimestamp: time.Now(),
		TraceTimestamp:    time.Now(),
		Request:           "",
		Response:          "",
		ServicePath:       "",
		Hop:               0,
		Tags:              make(map[string]string),
	}
}

func (l *Trace) SaveToMapStringList() (map[string][]string, error) {
	result := make(map[string][]string)

	result["x-trace-uid"] = []string{l.UID}
	result["x-trace-hop"] = []string{fmt.Sprintf("%d", l.Hop)}
	result["x-trace-request-timestamp"] = []string{l.RequestTimestamp.Format(time.RFC3339)}
	result["x-trace-response-timestamp"] = []string{l.ResponseTimestamp.Format(time.RFC3339)}
	result["x-trace-timestamp"] = []string{l.TraceTimestamp.Format(time.RFC3339)}
	result["x-trace-request"] = []string{l.Request}
	result["x-trace-response"] = []string{l.Response}
	result["x-trace-total-request-time"] = []string{l.TotalRequestTime}
	result["x-trace-service-path"] = []string{l.UID}

	for key, value := range l.Tags {
		result[key] = []string{value}
	}

	return result, nil
}

func (l *Trace) LoadFromMapStringList(msl map[string][]string) error {
	for name, values := range msl {
		if len(values) == 0 {
			continue
		}
		value := values[0]
		switch strings.ToLower(name) {
		case "x-trace-uid":
			l.UID = value
		case "x-trace-request-timestamp":
			if t, err := time.Parse(time.RFC3339, value); err == nil {
				l.RequestTimestamp = t
			} else {
				l.RequestTimestamp = time.Now()
			}
		case "x-trace-response-timestamp":
			if t, err := time.Parse(time.RFC3339, value); err == nil {
				l.ResponseTimestamp = t
			} else {
				l.ResponseTimestamp = time.Now()
			}
		case "x-trace-timestamp":
			if t, err := time.Parse(time.RFC3339, value); err == nil {
				l.TraceTimestamp = t
			} else {
				l.TraceTimestamp = time.Now()
			}
		case "x-trace-hop":
			if i, err := strconv.Atoi(value); err == nil {
				l.Hop = i
			} else {
				l.Hop = 0
			}
		case "x-trace-request":
			l.Request = value
		case "x-trace-response":
			l.Response = value
		case "x-trace-service-path":
			l.ServicePath = value
		case "x-trace-total-request-time":
			l.TotalRequestTime = value
		default:
			l.Tags[name] = value
		}
	}

	return nil
}

func (l *Trace) formatTags() string {
	if len(l.Tags) == 0 {
		return ""
	}

	var builder strings.Builder

	for key, value := range l.Tags {
		builder.WriteString(key)
		builder.WriteString(":")
		builder.WriteString(value)
		builder.WriteString(",")
	}

	result := builder.String()
	if len(result) > 0 {
		result = result[:len(result)-1]
	}

	result += ";"

	return result
}

func (l *Trace) IncrementHop(serviceName string, totalRequestTimeUs int64) {
	l.Hop = l.Hop + 1
	l.ServicePath += " => " + "[" + strconv.Itoa(l.Hop) + "] " + serviceName
	l.TotalRequestTime += " => " + "[" + strconv.Itoa(l.Hop) + "] " + strconv.FormatInt(totalRequestTimeUs, 10)
}

func (l *Trace) LoadFromHttpRequest(r *http.Request) error {
	return l.LoadFromMapStringList(r.Header)
}

func (l *Trace) SaveToHttpResponse(w *http.Response) {
	for key, value := range l.Tags {
		w.Header.Set(key, value)
	}
}

func (l *Trace) SaveToHttpResponseWriter(w http.ResponseWriter) {
	for key, value := range l.Tags {
		w.Header().Set(key, value)
	}
}

func (l *Trace) ToString() string {
	return fmt.Sprintf(`%s - [%s]
Request: %s 
Response: %s
Hop: %d
Tags: %s`,
		l.TraceTimestamp.Format("2006-01-02 15:04:05"),
		l.UID,
		l.Request,
		l.Response,
		l.Hop,
		l.formatTags())
}
func fallbackErrorLog(message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf(fmt.Sprintf("%s - [FALLBACK] : %s\n", timestamp, message))
}

func fallbackLog(m Trace) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf(fmt.Sprintf("%s - [FALLBACK] [%s] : %s\n", timestamp, m.formatTags(), m.ToString()))
}
