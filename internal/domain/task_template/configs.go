package tasktemplate

type DailyConfig struct {
	Interval int      `json:"interval"`
	Times    []string `json:"times"`
}

type MonthlyConfig struct {
	DayOfMonth int      `json:"day_of_month"`
	Times      []string `json:"times"`
}

type SpecificDatesConfig struct {
	Date  string   `json:"date"`
	Times []string `json:"times"`
}

type DayParityConfig struct {
	Parity string   `json:"parity"`
	Times  []string `json:"times"`
}
