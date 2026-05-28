package billing

type Metric struct {
	key   string
	value int64
}

func NewMetric(key string, value int64) Metric {
	return Metric{
		key:   key,
		value: value,
	}
}

func (m Metric) Key() string {
	return m.key
}

func (m Metric) Value() int64 {
	return m.value
}
