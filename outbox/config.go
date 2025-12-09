package outbox

import (
	"fmt"
	"math"
	"strings"
	"time"
)

type Config struct {
	// Limit of events to receive, min = 1, Max = 1000
	Limit int

	// Num of workers for publish events, min = 1, max = 32
	// Workers int

	// Event processing interval, min = 100ms
	Interval time.Duration

	// Event reserve duration, min = 15s
	ReserveDuration time.Duration

	// Timeout for processing method, min = 200 ms
	ProcessTimeout time.Duration

	//Header event_id key
	HeaderEventID string
	//Header event_type key
	HeaderEventType string
}

var SmallBatchConfig = Config{
	Limit:           50,
	Interval:        250 * time.Millisecond,
	ReserveDuration: 15 * time.Second,
	ProcessTimeout:  300 * time.Millisecond,
}

var MediumBatchConfig = Config{
	Limit:           200,
	Interval:        450 * time.Millisecond,
	ReserveDuration: 30 * time.Second,
	ProcessTimeout:  530 * time.Millisecond,
}

var LargeBatchConfig = Config{
	Limit:           500,
	Interval:        620 * time.Second,
	ReserveDuration: 45 * time.Second,
	ProcessTimeout:  720 * time.Second,
}

func validateConfig(cfg *Config) error {
	const (
		minLimit          = 1
		maxLimit          = 1000
		minInterval       = 100 * time.Millisecond
		minReserve        = 15 * time.Second
		minProcessTimeout = 200 * time.Millisecond
	)

	var errs []string

	if cfg.Limit < minLimit || cfg.Limit > maxLimit {
		errs = append(errs, fmt.Sprintf("limit must be in [%d;%d]", minLimit, maxLimit))
	}

	if cfg.Interval < minInterval {
		errs = append(errs, fmt.Sprintf("interval must be >= %s", minInterval))
	}

	if cfg.ReserveDuration < minReserve {
		errs = append(errs, fmt.Sprintf("reserveDuration must be >= %s", minReserve))
	}

	if cfg.ProcessTimeout < minProcessTimeout {
		errs = append(errs, fmt.Sprintf("processTimeout must be >= %s", minProcessTimeout))
	}

	if cfg.HeaderEventID == "" {
		errs = append(errs, "headerEventID required")
	}

	if cfg.HeaderEventType == "" {
		errs = append(errs, "headerEventType required")
	}

	if len(errs) > 0 {
		return fmt.Errorf("invalid config:\n - %s", strings.Join(errs, "\n - "))
	}

	return nil
}

type FlushConfug struct {
	MaxMessages int
	Frequency   time.Duration
}

func (c Config) GetKafkaFlushConfig() FlushConfug {

	maxMessages := max(int(math.Floor(float64(c.Limit)*0.4)), 1)

	frequency := time.Duration(float64(c.ProcessTimeout) * 0.3)

	return FlushConfug{
		MaxMessages: maxMessages,
		Frequency:   frequency,
	}

}
