package tracing

import (
	"fmt"
	"time"
)

type Fallback struct {
}

func NewFallback() *Fallback {
	return &Fallback{}
}

func (lf *Fallback) Log(m Trace) error {
	fmt.Printf("%s - [%s] : %s\n", time.Now().Format("2006-01-02 15:04:05"), "FALLBACK", m.ToString())
	return nil
}
