package rotatelogs

import (
	"fmt"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
)

func TestGenFilename(t *testing.T) {
	// Mock time
	ts := []time.Time{
		time.Time{},
		(time.Time{}).Add(24 * time.Hour),
	}

	for _, xt := range ts {
		rl, err := New(
			"/path/to/%Y/%m/%d",
			WithClock(clockwork.NewFakeClockAt(xt)),
		)
		if !assert.NoError(t, err, "New should succeed") {
			return
		}

		defer rl.Close()

		fn := rl.genFilename()
		expected := fmt.Sprintf("/path/to/%04d/%02d/%02d",
			xt.Year(),
			xt.Month(),
			xt.Day(),
		)

		if !assert.Equal(t, expected, fn) {
			return
		}
	}
}

func TestWithLocation(t *testing.T) {
	// Not really a test, but well...
	loc, _ := time.LoadLocation("Asia/Tokyo")
	opt := WithLocation(loc)
	var rl RotateLogs
	opt.Configure(&rl)
	t.Logf("%s", rl.clock.Now())
}

