package tracing

type Configuration struct {
	UseConsole   bool
	UseNATS      bool
	NatsURL      string
	NatsUsername string
	NatsPassword string
}

func NewConfiguration() *Configuration {
	return &Configuration{}
}

func (c *Configuration) Init() ITracing {
	mt := NewMultiTracer()
	mt.tracers = make([]ITracer, 0)

	//mt.tracers = append(mt.tracers, NewFallback())

	if c.UseConsole {
		mt.tracers = append(mt.tracers, NewConsole())
	}

	if c.UseNATS {
		if c.NatsUsername != "" && c.NatsPassword != "" {
			if natsLogger, err := NewMetricsNATSWithAuth(c.NatsURL, c.NatsUsername, c.NatsPassword); err == nil {
				mt.tracers = append(mt.tracers, natsLogger)
			}
		} else {
			if natsLogger, err := NewMetricsNATS(c.NatsURL); err == nil {
				mt.tracers = append(mt.tracers, natsLogger)
			}
		}
	}

	return mt
}
