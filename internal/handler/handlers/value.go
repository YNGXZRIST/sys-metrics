package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sys-metrics/internal/model/metrics"
	svm "sys-metrics/internal/service/metrics"
	"sys-metrics/internal/service/responsewriter"
	"sys-metrics/pkg/stringsparser"
)

func ValueHandler(w http.ResponseWriter, r *http.Request) {
	metricType := r.PathValue("type")
	name := r.PathValue("name")
	metricType = strings.ToLower(metricType)
	name = stringsparser.Capitalize(name)
	switch metricType {
	case metrics.MTypeCounter:
		v, err := svm.Counters().Get(name)
		if err != nil {
			fmt.Println(err)
			responsewriter.WriteNotFound(w)
			return
		}
		responsewriter.WriteSuccessStatus(w)
		w.Write([]byte(strconv.FormatInt(*v.Delta, 10)))
		return
	case metrics.MTypeGauge:
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
