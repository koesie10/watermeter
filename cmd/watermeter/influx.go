package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/koesie10/watermeter/influx"
	"github.com/koesie10/watermeter/watermeter"
	"github.com/spf13/cobra"
)

var influxOptions = client.HTTPConfig{}

var additionalInfluxOptions = struct {
	Database        string
	RetentionPolicy string

	WaterMeasurement string

	Tags []string

	DisableUpload bool
}{}

var influxCmd = &cobra.Command{
	Use:   "influx",
	Short: "sends an observation to InfluxDB when the value changes",
	RunE: func(cmd *cobra.Command, args []string) error {
		tags := make(map[string]string)

		for _, v := range additionalInfluxOptions.Tags {
			parts := strings.SplitN(v, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid tag %q", v)
			}

			tags[parts[0]] = parts[1]
		}

		c, err := client.NewHTTPClient(influxOptions)
		if err != nil {
			return fmt.Errorf("failed to connect to InfluxDB: %w", err)
		}
		defer c.Close()

		if !additionalInfluxOptions.DisableUpload {
			if _, _, err := c.Ping(5 * time.Second); err != nil {
				return fmt.Errorf("failed to ping InfluxDB: %w", err)
			}
		}

		pin, err := OpenPin()
		if err != nil {
			return fmt.Errorf("failed to open pin: %w", err)
		}

		wm, err := watermeter.New(pin)
		if err != nil {
			return fmt.Errorf("failed to open water meter: %w", err)
		}

		if err := wm.RegisterWatcher(func(value int64) {
			if jsonOutput {
				data, err := json.Marshal(map[string]int64{"value": value})
				if err != nil {
					log.Println(err)
					return
				}

				fmt.Println(string(data))
			}

			ep, err := influx.NewWaterPoint(time.Now(), value, additionalInfluxOptions.WaterMeasurement, tags)
			if err != nil {
				log.Println(err)
				return
			}

			bp, err := client.NewBatchPoints(client.BatchPointsConfig{
				Database:        additionalInfluxOptions.Database,
				RetentionPolicy: additionalInfluxOptions.RetentionPolicy,
			})
			if err != nil {
				log.Println(err)
				return
			}

			bp.AddPoint(ep)

			if additionalInfluxOptions.DisableUpload {
				for _, p := range bp.Points() {
					fmt.Println(p.PrecisionString("ns"))
				}
			} else {
				if err := c.Write(bp); err != nil {
					log.Println(err)
					return
				}
			}
		}); err != nil {
			return fmt.Errorf("failed to register watcher: %w", err)
		}

		return WaitForExit()
	},
}

func init() {
	rootCmd.AddCommand(influxCmd)
	influxCmd.Flags().StringVar(&influxOptions.Addr, "influx-addr", "http://localhost:8086", "InfluxDB address")
	influxCmd.Flags().StringVar(&influxOptions.Username, "influx-username", "", "InfluxDB username")
	influxCmd.Flags().StringVar(&influxOptions.Password, "influx-password", "", "InfluxDB password")
	influxCmd.Flags().DurationVar(&influxOptions.Timeout, "influx-timeout", 10*time.Second, "InfluxDB timeout")

	influxCmd.Flags().StringVar(&additionalInfluxOptions.Database, "influx-database", "watermeter", "InfluxDB database")
	influxCmd.Flags().StringVar(&additionalInfluxOptions.RetentionPolicy, "influx-retention-policy", "", "InfluxDB retention policy. Leave empty for default.")
	influxCmd.Flags().StringVar(&additionalInfluxOptions.WaterMeasurement, "influx-water-measurement", "watermeter_water", "InfluxDB measurement for water")

	influxCmd.Flags().StringSliceVar(&additionalInfluxOptions.Tags, "influx-tags", []string{}, "InfluxDB tags in key=value format")

	influxCmd.Flags().BoolVar(&additionalInfluxOptions.DisableUpload, "disable-upload", false, "if upload is disabled, then all points will be written to stdout")
}
