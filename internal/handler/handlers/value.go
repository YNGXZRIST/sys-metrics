package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	collector "sys-metrics/internal/agent"
	"sys-metrics/internal/common"
	svm "sys-metrics/internal/service/metrics"
	"sys-metrics/internal/service/responsewriter"
)

func ValueHandler(w http.ResponseWriter, r *http.Request) {
	metricType := r.PathValue("type")
	name := r.PathValue("name")
	metricType = strings.ToLower(metricType)
	name = collector.GetMetricType(name)
	switch metricType {
	case common.Counter:
		v, err := svm.Counters().Get(name)
		if err != nil {
			fmt.Println(err)
			responsewriter.WriteNotFound(w)
			return
		}
		responsewriter.WriteSuccessStatus(w)
		w.Write([]byte(strconv.FormatInt(*v.Delta, 10)))
		return
	case common.Gauge:
		v, err := svm.Gauges().Get(name)
		if err != nil {
			fmt.Println(err)
			responsewriter.WriteNotFound(w)
			return
		}
		responsewriter.WriteSuccessStatus(w)
		w.Write([]byte(strconv.FormatFloat(*v.Value, 'f', -1, 64)))
		return
	default:
		responsewriter.WriteBadRequest(w)
		return
	}
}
