package influx

import (
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

func NewWaterPoint(t time.Time, value int64, measurement string, tags map[string]string) (*client.Point, error) {
	tags = copyTags(tags)

	fields := make(map[string]interface{})
	fields["value"] = value

	return client.NewPoint(measurement, tags, fields, t)
}

func copyTags(tags map[string]string) map[string]string {
	if len(tags) == 0 {
		return tags
	}
	result := make(map[string]string, len(tags))
	for k, v := range tags {
		result[k] = v
	}
	return result
}
