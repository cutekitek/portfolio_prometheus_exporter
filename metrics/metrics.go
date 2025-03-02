package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var ()

type MetricsManager interface {
	UpdateAccountMoneyValue(currency string, value float64)
	UpdateAccountShareCount(shareName string, value float64)
	UpdateAccountOpenedPositions(shareName string, value float64)
	UpdateAccountRelativeYield(value float64)
	UpdateShareRelativeYield(share string, value float64)
	UpdateSharePrice(share string, price float64)
}

type manager struct {
	registry *prometheus.Registry
	exchange string

	accValue   *prometheus.GaugeVec
	shareCount *prometheus.GaugeVec
	openedPos  *prometheus.GaugeVec
	relativeYield *prometheus.GaugeVec
	shareRelYield *prometheus.GaugeVec
	sharePrice *prometheus.GaugeVec
}

func NewMetricsManager(registry *prometheus.Registry, exchange string) MetricsManager {
	m := &manager{registry: registry, exchange: exchange}
	m.register()
	
	return m
}

func (m *manager) UpdateAccountMoneyValue(currency string, value float64) {
	m.accValue.WithLabelValues(m.exchange, currency).Set(value)
}

func (m *manager) UpdateAccountOpenedPositions(shareName string, value float64) {
	m.openedPos.WithLabelValues(m.exchange, shareName).Set(value)
}

func (m *manager) UpdateAccountShareCount(shareName string, value float64) {
	m.shareCount.WithLabelValues(m.exchange, shareName).Set(value)
}

func (m *manager) UpdateAccountRelativeYield(value float64) {
	m.relativeYield.WithLabelValues(m.exchange).Set(value)
}

func (m *manager) UpdateShareRelativeYield(share string, value float64) {
	m.shareRelYield.WithLabelValues(m.exchange, share).Set(value)
} 

func (m *manager) UpdateSharePrice(share string, price float64) {
	m.sharePrice.WithLabelValues(m.exchange, share).Set(price)
}

func (m *manager) register() {
	m.accValue = promauto.With(m.registry).NewGaugeVec(prometheus.GaugeOpts{
		Name:        "money_value",
		Help:        "money value of all account shares",
		ConstLabels: prometheus.Labels{},
	}, []string{"exchange", "currency"})
	m.shareCount = promauto.With(m.registry).NewGaugeVec(prometheus.GaugeOpts{
		Name: "shares_count",
		Help: "amount of shares in portfolio",
	}, []string{"exchange", "share"})
	m.openedPos = promauto.With(m.registry).NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "opened_positions",
			Help: "qwd",
		}, []string{"exchange", "currency"},
	)
	m.relativeYield = promauto.With(m.registry).NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "total_relative_yield",
			Help: "relative yield of portfolio ",
		}, []string{"exchange"},
	)
	m.shareRelYield = promauto.With(m.registry).NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "share_relative_yield",
			Help: "relative yield of portfolio ",
		}, []string{"exchange", "share"},
	)
	m.sharePrice = promauto.With(m.registry).NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "share_price",
			Help: "share price",
		}, []string{"exchange", "share"},
	)
}
