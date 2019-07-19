package collector

import (
	"net/url"

	"github.com/prometheus/client_golang/prometheus"
)

func constLabelsFromURL(url *url.URL) prometheus.Labels {
	u := *url
	u.User = nil
	return map[string]string{
		"cluster_url": u.String(),
	}
}
