package amplitude

import (
	"os"

	"github.com/amplitude/analytics-go/amplitude"
)

type Amplitude struct {
	Client      amplitude.Client
	environment string
}

func Initialize() Amplitude {
	config := amplitude.NewConfig(os.Getenv("AMPLITUDE_API_KEY"))
	config.FlushQueueSize = 100
	config.FlushInterval = 5000

	return Amplitude{
		Client:      amplitude.NewClient(config),
		environment: os.Getenv("ENVIRONMENT"),
	}
}

func (a Amplitude) TrackEvent(userId string, eventType string, eventProperties map[string]interface{}) {
	if a.environment != "production" {
		return
	}

	a.Client.Track(amplitude.Event{
		UserID:          userId,
		EventType:       eventType,
		EventProperties: eventProperties,
	})
}
