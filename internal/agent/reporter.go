package agent

import (
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
)

type Reporter struct {
	serverAddr string
	logger     *log.Logger
}

func NewReporter(serverAddr string, logger *log.Logger) *Reporter {
	return &Reporter{serverAddr, logger}
}
func (r *Reporter) Send(c Collector) error {
	for t, metrics := range c.metrics {
		for m, v := range metrics {
			err := r.sendMetricToServer(t, m, v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func (r *Reporter) sendMetricToServer(metric, name string, value float64) error {
	if metric == "" {
		return errors.New("empty metric type")
	}
	if name == "" {
		return errors.New("empty metric name")
	}

	v := r.ConvertMetricValue(metric, value)
	url := r.BuildUpdateURL(metric, name, v)
	r.logger.Println(url)
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	_, err = io.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		return err
	}
	r.logger.Println(response.Status)
	r.logger.Println(response.StatusCode)
	return nil
}
func (r *Reporter) ConvertMetricValue(m string, v float64) string {
	var s string
	if m == "counter" {
		s = strconv.FormatInt(int64(v), 10)
	} else {
		s = strconv.FormatFloat(v, 'f', -1, 64)
	}
	return s
}
func (r *Reporter) BuildUpdateURL(m, n, v string) string {
	return r.serverAddr + "/update/" + m + "/" + n + "/" + v
}
