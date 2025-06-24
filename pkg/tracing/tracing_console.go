package tracing

import (
	"fmt"
	"time"
)

type Console struct{}

func NewConsole() *Console {
	return &Console{}
}

func (c Console) Log(m Trace) error {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf(fmt.Sprintf("%s - [CONSOLE] [%s] : %s \n", timestamp, m.formatTags(), m.ToString()))
	return nil
}
