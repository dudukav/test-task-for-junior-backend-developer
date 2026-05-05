package tasktemplate

type DailyConfig struct {
	Interval 	int
	Times 		[]string
}

type MonthlyConfig struct {
	DayofMonth 	int
	Times 		[]string
}

type SpecificDatesConfig struct {
	Date 		string
	Times 		[]string
}

type DayParityConfig struct {
	Parity 		string
	Times 		[]string
}