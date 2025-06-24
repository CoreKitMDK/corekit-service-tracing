package tracing

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

//TraceHttpResponseForRequest(r *http.Request, w http.Response) ITracing
//TraceHttpResponseWriterForRequest(r *http.Request, w http.ResponseWriter) ITracing
//Trace(m *Trace) ITracing

type ITracing interface {
	TraceHttpRequest(r *http.Request) ITracing
	TraceHttpResponse(r *http.Response) ITracing
	TraceHttpResponseWriter(w http.ResponseWriter) ITracing

	Start() ITracing
	End() ITracing
}

type ITracer interface {
	Log(m Trace) error
}

// Usage should be :
// trace := multiTracer.TraceHttpRequest(x).Start()
// trace.TraceHttpResponseWriter(w).Stop()

type MultiTracer struct {
	hostname     string
	tracers      []ITracer
	currentTrace *Trace
	startTime    time.Time
	endTime      time.Time
}

func NewMultiTracer() *MultiTracer {
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Printf("Error getting hostname: %v\n", err)
		hostname = "unknown"
	}

	return &MultiTracer{
		hostname:     hostname,
		tracers:      make([]ITracer, 0),
		currentTrace: nil,
		startTime:    time.Now(),
		endTime:      time.Now(),
	}
}

func NewMultiTracerForTrace(t *Trace) *MultiTracer {
	mt := NewMultiTracer()
	mt.currentTrace = t
	return mt
}

func (mt *MultiTracer) TraceHttpRequest(r *http.Request) ITracing {
	ct := NewTrace()
	err := ct.LoadFromHttpRequest(r)
	if err != nil {
		fallbackErrorLog("Failed to load from http request on host : " + mt.hostname)
		return mt
	}

	nmt := NewMultiTracerForTrace(ct)
	nmt.startTime = time.Now()
	nmt.tracers = mt.tracers

	return nmt
}

func (mt *MultiTracer) TraceHttpResponse(r *http.Response) ITracing {
	if mt.currentTrace == nil {
		fallbackErrorLog("Failed to trace http response on host : " + mt.hostname)
	}

	mt.currentTrace.SaveToHttpResponse(r)

	return mt
}

func (mt *MultiTracer) TraceHttpResponseWriter(w http.ResponseWriter) ITracing {
	if mt.currentTrace == nil {
		fallbackErrorLog("Failed to trace http response on host : " + mt.hostname)
		return mt
	}

	mt.currentTrace.SaveToHttpResponseWriter(w)

	return mt
}

func (mt *MultiTracer) Start() ITracing {
	if mt.currentTrace == nil {
		fallbackErrorLog("Failed to trace http response on host : " + mt.hostname)
		return mt
	}

	mt.startTime = time.Now()
	mt.currentTrace.RequestTimestamp = mt.startTime
	return mt
}

func (mt *MultiTracer) End() ITracing {
	if mt.currentTrace == nil {
		fallbackErrorLog("Failed to trace http response on host : " + mt.hostname)
		return mt
	}

	mt.endTime = time.Now()
	mt.currentTrace.ResponseTimestamp = mt.endTime
	mt.currentTrace.TraceTimestamp = mt.endTime
	mt.currentTrace.IncrementHop(mt.hostname, mt.endTime.Sub(mt.startTime).Nanoseconds()/1000)
	mt.writeToTracers()
	return mt
}

func (mt *MultiTracer) writeToTracers() {
	if mt.currentTrace == nil {
		fallbackErrorLog("Failed to trace http response on host : " + mt.hostname)
		return
	}

	for _, tracer := range mt.tracers {
		go func() {
			if err := tracer.Log(*mt.currentTrace); err != nil {
				fallbackErrorLog("Error writing to tracer:" + err.Error())
			}
		}()
	}
}
