package httpheader

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Event defines a Google Calendar hook event type
type Event string

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
			plrun := GoogleCalendarPayload{
				ChannelID:     "channel-ID-value",
				ChannelToken:  "channel-token-value",
				ResourceID:    "identifier-for-the-watched-resource",
				ResourceURI:   "version-specific-URI-of-the-watched-resource",
				MessageNumber: 1,
			}
			plrun.ChannelExpiration, _ = time.Parse(http.TimeFormat, "Tue, 19 Nov 2013 01:13:52 GMT")
			plrun.ResourceState = string(tt.args.e)
			gcp := GoogleCalendarPayload{}
			err := Decode(getHeader(tt.args.e), &gcp)
			if (err != nil) != tt.wantErr {
				t.Errorf("%d. Decode() error = %+v, wantErr %+v", i, err, tt.wantErr)
			}
			assert.Equal(t, plrun, gcp)
		})
	}
}

type DecodedArgs []string

func (m *DecodedArgs) DecodeHeader(header http.Header, tagKey string) error {
	baseKey := http.CanonicalHeaderKey(tagKey)
	keyMatch := regexp.MustCompile(fmt.Sprintf(`^%s\.\d+$`, baseKey))
	var args DecodedArgs
	for k := range header {
		if keyMatch.MatchString(http.CanonicalHeaderKey(k)) {
			args = append(args, header.Get(k))
		}
	}
	// TODO: sort args by id
	sort.Strings(args)
	if len(args) > 0 {
		*m = args
	}
	return nil
}

func TestDecodeHeader_Unmarshaler(t *testing.T) {
	type ArgStruct struct {
		Args DecodedArgs `header:"Arg"`
	}
	input := http.Header{
		"Arg.0": []string{"a"},
		"Arg.1": []string{"b"},
		"Arg.2": []string{"c"},
	}
	want := ArgStruct{
		Args: []string{"a", "b", "c"},
	}
	var got ArgStruct

	err := Decode(input, &got)
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestDecodeHeader_UnmarshalerWithNilPointer(t *testing.T) {
	s := struct {
		Args *EncodedArgs `header:"Arg"`
	}{}
	got, err := Header(s)
	assert.NoError(t, err)

	want := http.Header{}
	assert.Equal(t, want, got)
}

type simpleStruct struct {
	Foo string
}

type fullTypeStruct struct {
	unExport          string
	UnExportTwo       string `header:"-"`
	Bool              bool   `header:"Bool"`
	BoolInt           bool   `header:"Bool-Int,int"`
	String            string
	StringEmpty       string `header:"String-Empty"`
	StringEmptyIgnore string `header:"String-Empty-Ignore,omitempty"`
	Uint              uint
	Uint64            uint64
	Uint8             uint8
	Uint16            uint16
	Uint32            uint32
	Int               int
	Int64             int64
	Int8              int8
	Int16             int16
	Int32             int32
	Float32           float32
	Float64           float64
	Slice             []string
	SliceTwo          []int `header:"Slice-Two"`
	Array             [3]string
	ArrayTwo          [2]int `header:"Array-Two"`
	Interface         interface{}
	Time              time.Time
	TimeUnix          time.Time `header:"Time-Unix,unix"`
	Point             *string
	Args              DecodedArgs `header:"Arg"`
	Foo               simpleStruct
}

func TestDecodeHeader_2(t *testing.T) {
	timeV := time.Date(2000, 1, 1, 12, 34, 56, 0, time.UTC)
	timeS := "Sat, 01 Jan 2000 12:34:56 GMT"
	timeU := "946730096"
	h := http.Header{
		"UnExportTwo":  []string{"foo"},
		"UnExport-Two": []string{"foo"},
		"Bool":         []string{"true"},
		"Bool-Int":     []string{"1"},
		"String":       []string{"foobar"},
		"String-Empty": []string{""},
		"Uint":         []string{"2"},
		"Uint64":       []string{"3"},
		"Uint8":        []string{"4"},
		"Uint16":       []string{"5"},
		"Uint32":       []string{"6"},
		"Int":          []string{"7"},
		"Int64":        []string{"8"},
		"Int8":         []string{"9"},
		"Int16":        []string{"10"},
		"Int32":        []string{"11"},
		"Float32":      []string{"12.2"},
		"Float64":      []string{"13.2"},
		"Slice":        []string{"a", "b", "c"},
		"Slice-Two":    []string{"1", "2", "3"},
		"Array":        []string{"a", "b", "c"},
		"Array-Two":    []string{"1", "2", "3"},
		"Interface":    []string{"foo", "bar"},
		"Time":         []string{timeS},
		"Time-Unix":    []string{timeU},
		"Point":        []string{"foo"},
		"Arg.0":        []string{"a"},
		"Arg.1":        []string{"b"},
		"Arg.2":        []string{"c"},
		"Foo":          []string{"bar"},
	}
	want := fullTypeStruct{
		unExport:          "",
		UnExportTwo:       "",
		Bool:              true,
		BoolInt:           true,
		String:            "foobar",
		StringEmpty:       "",
		StringEmptyIgnore: "",
		Uint:              2,
		Uint64:            3,
		Uint8:             4,
		Uint16:            5,
		Uint32:            6,
		Int:               7,
		Int64:             8,
		Int8:              9,
		Int16:             10,
		Int32:             11,
		Float32:           12.2,
		Float64:           13.2,
		Slice:             []string{"a", "b", "c"},
		SliceTwo:          []int{1, 2, 3},
		Array:             [3]string{"a", "b", "c"},
		ArrayTwo:          [2]int{1, 2},
		Interface:         interface{}([]string{"foo", "bar"}),
		Time:              timeV,
		TimeUnix:          timeV,
		Point:             stringPoint("foo"),
		Args:              []string{"a", "b", "c"},
		Foo:               simpleStruct{Foo: "bar"},
	}
	var got fullTypeStruct
	err := Decode(h, &got)
	assert.NoError(t, err)
	assert.EqualValues(t, want, got)
}

func stringPoint(s string) *string {
	return &s
}

func Test_fillValues_errors(t *testing.T) {
	type args struct {
		sv     reflect.Value
		opts   tagOptions
		valArr []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Uint",
			args: args{
				sv:     reflect.New(reflect.TypeOf(uint(3))),
				opts:   tagOptions{},
				valArr: []string{"a"},
			},
			wantErr: true,
		},
		{
			name: "Uint64",
			args: args{
				sv:     reflect.New(reflect.TypeOf(uint64(3))),
				opts:   tagOptions{},
				valArr: []string{"a"},
			},
			wantErr: true,
		},
		{
			name: "Uint8",
			args: args{
				sv:     reflect.New(reflect.TypeOf(uint8(3))),
				opts:   tagOptions{},
				valArr: []string{"a"},
			},
			wantErr: true,
		},
		{
			name: "Uint16",
			args: args{
				sv:     reflect.New(reflect.TypeOf(uint16(3))),
				opts:   tagOptions{},
				valArr: []string{"a"},
			},
			wantErr: true,
		},
		{
			name: "Uint32",
			args: args{
				sv:     reflect.New(reflect.TypeOf(uint32(3))),
				opts:   tagOptions{},
				valArr: []string{"a"},
			},
			wantErr: true,
		},
		{
			name: "int",
			args: args{
				sv:     reflect.New(reflect.TypeOf(int(3))),
				opts:   tagOptions{},
				valArr: []string{"a"},
			},
			wantErr: true,
		},
		{
			name: "int64",
			args: args{
				sv:     reflect.New(reflect.TypeOf(int64(3))),
				opts:   tagOptions{},
				valArr: []string{"a"},
			},
			wantErr: true,
		},
		{
			name: "int8",
			args: args{
				sv:     reflect.New(reflect.TypeOf(int8(3))),
				opts:   tagOptions{},
				valArr: []string{"a"},
			},
			wantErr: true,
		},
		{
			name: "int16",
			args: args{
				sv:     reflect.New(reflect.TypeOf(int16(3))),
				opts:   tagOptions{},
				valArr: []string{"a"},
			},
			wantErr: true,
		},
		{
			name: "int32",
			args: args{
				sv:     reflect.New(reflect.TypeOf(int32(3))),
				opts:   tagOptions{},
				valArr: []string{"a"},
			},
			wantErr: true,
		},
		{
			name: "float32",
			args: args{
				sv:     reflect.New(reflect.TypeOf(float32(3))),
				opts:   tagOptions{},
				valArr: []string{"a"},
			},
			wantErr: true,
		},
		{
			name: "float64",
			args: args{
				sv:     reflect.New(reflect.TypeOf(float64(3))),
				opts:   tagOptions{},
				valArr: []string{"a"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := fillValues(tt.args.sv, tt.args.opts, tt.args.valArr); (err != nil) != tt.wantErr {
				t.Errorf("fillValues() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
