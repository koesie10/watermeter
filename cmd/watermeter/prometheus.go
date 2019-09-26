package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/koesie10/watermeter/watermeter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

var promOptions = struct {
	Addr string
}{}

var prometheusCmd = &cobra.Command{
	Use:   "prometheus",
	Short: "Start a web server and serve the data of the water meter on /metrics",
	RunE: func(cmd *cobra.Command, args []string) error {
		tags := make(map[string]string)

		for _, v := range additionalInfluxOptions.Tags {
			parts := strings.SplitN(v, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid tag %q", v)
			}

			tags[parts[0]] = parts[1]
		}

		pin, err := OpenPin()
		if err != nil {
			return fmt.Errorf("failed to open pin: %w", err)
		}

		wm, err := watermeter.New(pin)
		if err != nil {
			return fmt.Errorf("failed to open water meter: %w", err)
		}

		registry := prometheus.NewRegistry()

		value := prometheus.NewCounterFunc(prometheus.CounterOpts{
			Name:      "value_liters",
			Help:      "Actual water value registered in the current run of the program",
			Subsystem: "water",
			Namespace: "watermeter",
		}, func() float64 {
			return float64(wm.NumRises())
		})

		registry.MustRegister(value)

		http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

		go func() {
			log.Fatal(http.ListenAndServe(promOptions.Addr, nil))
		}()

		return WaitForExit()
	},
}

func init() {
	rootCmd.AddCommand(prometheusCmd)
	prometheusCmd.Flags().StringVar(&promOptions.Addr, "addr", ":8888", "Web server address")
}
