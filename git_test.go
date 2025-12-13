package gomake

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGeneratedVersionTime(t *testing.T) {
	versionTime, _ := GeneratedVersionTime(time.Date(2020, 12, 12, 0, 0, 0, 0, time.UTC))
	matched, _ := regexp.Match(`1.20201212.0-H[a-z0-9]{7,8}`, []byte(versionTime))
	assert.True(t, matched)
}
