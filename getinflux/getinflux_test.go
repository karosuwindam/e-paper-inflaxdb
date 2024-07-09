package getinflux

import (
	"fmt"
	"testing"
	"time"
)

func TestString(t *testing.T) {
	influx := influxDBPass{url: "http://test:01", db: "senser", table: "test_sanple"}
	fmt.Println(createReadUrlData(influx, 1, "test"))
	fmt.Println(createReadUrlData(influx, time.Second, "test"))
}
