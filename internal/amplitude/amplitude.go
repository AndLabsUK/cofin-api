package amplitude

import (
	"github.com/amplitude/analytics-go/amplitude"
	"os"
)

type Amplitude struct {
	Client amplitude.Client
}

func Initialize() Amplitude {
	config := amplitude.NewConfig(os.Getenv("AMPLITUDE_API_KEY"))
	config.FlushQueueSize = 100
	config.FlushInterval = 5000

	return Amplitude{
		Client: amplitude.NewClient(config),
	}
}

func (a Amplitude) TrackEvent(userId string, eventType string, eventProperties map[string]interface{}) {
	a.Client.Track(amplitude.Event{
		UserID:          userId,
		EventType:       eventType,
		EventProperties: eventProperties,
	})
}
