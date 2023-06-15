package security

import (
	"testing"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
)

func TestErrGenID(t *testing.T) {
	is := assert.New(t)
	errs := make(chan error, 100)
	for i := 0; i < cap(errs); i++ {
		if uid, err := GenID(); err == nil {
			parseUlid := ulid.MustParse(uid)
			is.Equal(ulid.Timestamp(now), parseUlid.Time())
			is.Equal(parseUlid.String(), uid)
		} else {
			errs <- err
		}
		errs <- nil
	}
	for i := 0; i < cap(errs); i++ {
		if err := <-errs; err != nil {
			is.Error(err)
		}
	}
}

func TestGenID(t *testing.T) {
	is := assert.New(t)
	type testCase struct {
		name         string
		actualTime   time.Time
		expectedTime time.Time
		pass         bool
	}
	cases := []testCase{
		{
			name:         "time now",
			actualTime:   time.Now(),
			expectedTime: time.Now(),
			pass:         true,
		},
		{
			name:         "time failed",
			actualTime:   time.Now(),
			expectedTime: time.Now().Add(time.Hour * 24),
			pass:         false,
		},
		{
			name:         "time nil",
			actualTime:   time.Time{},
			expectedTime: time.Now().Add(time.Hour * 24),
			pass:         false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			now = tt.actualTime
			uid, err := GenID()
			if err != nil {
				is.Error(err)
				return
			}
			parseUlid := ulid.MustParse(uid)
			if tt.pass {
				is.Equal(parseUlid.String(), uid)
				is.Equal(ulid.Timestamp(tt.expectedTime), parseUlid.Time())
			} else {
				is.Equal(parseUlid.String(), uid)
				is.NotEqual(ulid.Timestamp(tt.expectedTime), parseUlid.Time())
			}
		})
	}
}
