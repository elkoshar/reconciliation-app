package constants

import (
	"time"
)

func init() {
	JAKARTA_LOCATION, _ = time.LoadLocation("Asia/Jakarta")
}

var JAKARTA_LOCATION *time.Location
