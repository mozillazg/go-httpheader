package httpheader

import (
	"net/http"
	"reflect"
	"testing"
	"time"
)

// Event defines a Google Calendar hook event type
type Event string

var pl *GoogleCalendarPayload

// GoogleCalendar hook types
const (
	SyncEvent      Event = "sync"
	ExistsEvent    Event = "exists"
	NotExistsEvent Event = "not_exists"
)

// GoogleCalendarPayload a google calendar notice
// https://developers.google.com/calendar/v3/push
type GoogleCalendarPayload struct {
	ChannelID         string    `header:"X-Goog-Channel-ID"`
	ChannelToken      string    `header:"X-Goog-Channel-Token,omitempty"`
	ChannelExpiration time.Time `header:"X-Goog-Channel-Expiration,omitempty"`
	ResourceID        string    `header:"X-Goog-Resource-ID"`
	ResourceURI       string    `header:"X-Goog-Resource-URI"`
	ResourceState     string    `header:"X-Goog-Resource-State"`
	MessageNumber     int       `header:"X-Goog-Message-Number"`
}

func init() {
	pl = &GoogleCalendarPayload{
		ChannelID:     "channel-ID-value",
		ChannelToken:  "channel-token-value",
		ResourceID:    "identifier-for-the-watched-resource",
		ResourceURI:   "version-specific-URI-of-the-watched-resource",
		MessageNumber: 1,
	}
	pl.ChannelExpiration, _ = time.Parse(time.RFC1123, "Tue, 19 Nov 2013 01:13:52 GMT")

}

func getHeader(e Event) http.Header {
	h := http.Header{}
	h.Add("X-Goog-Channel-ID", "channel-ID-value")
	h.Add("X-Goog-Channel-Token", "channel-token-value")
	h.Add("X-Goog-Channel-Expiration", "Tue, 19 Nov 2013 01:13:52 GMT")
	h.Add("X-Goog-Resource-ID", "identifier-for-the-watched-resource")
	h.Add("X-Goog-Resource-URI", "version-specific-URI-of-the-watched-resource")
	h.Add("X-Goog-Message-Number", "1")
	h.Add("X-Goog-Resource-State", string(e))
	return h
}

func TestDecodeHeader(t *testing.T) {
	type args struct {
		e Event
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Google Calendar sync", args{SyncEvent}, false},
		{"Google Calendar exists", args{ExistsEvent}, false},
		{"Google Calendar no exists", args{NotExistsEvent}, false},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plrun := *pl
			plrun.ResourceState = string(tt.args.e)
			gcp := GoogleCalendarPayload{}
			err := DecodeHeader(getHeader(tt.args.e), &gcp)
			if (err != nil) != tt.wantErr {
				t.Errorf("%d. DecodeHeader() error = %+v, wantErr %+v", i, err, tt.wantErr)
			}
			if !reflect.DeepEqual(gcp, plrun) {
				t.Errorf("%d. DecodeHeader() does not work as expected, \ngot %+v \nwant %+v", i, gcp, plrun)
			}
		})
	}
}
