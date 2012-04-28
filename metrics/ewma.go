package metrics

const (
	M1_ALPHA  = 0.07995558537067670723530454779393039643764495849609 // 1 - math.Exp(-5 / 60.0)
	M5_ALPHA  = 0.01652854617838250828043555884505622088909149169922 // 1 - math.Exp(-5 / 60.0 / 5)
	M15_ALPHA = 0.00554015199510327072118798241717740893363952636719 // 1 - math.Exp(-5 / 60.0 / 15)
)

// An exponentially-weighted moving average.
//
// http://www.teamquest.com/pdfs/whitepaper/ldavg1.pdf - UNIX Load Average Part 1: How It Works
// http://www.teamquest.com/pdfs/whitepaper/ldavg2.pdf - UNIX Load Average Part 2: Not Your Average Average
type EWMA struct {
	interval  uint32  // exptected tick interval in seconds
	alpha     float64 // the smoothing constant
	uncounted float64
	rate      float64
}

func NewEWMA(interval uint32, alpha float64) *EWMA {
	return &EWMA{interval, alpha, 0.0, 0.0}
}

func (self *EWMA) Update(value float64) {
	self.uncounted += value
}

func (self *EWMA) Tick() {
	count := self.uncounted
	self.uncounted = 0
	instantRate := count / float64(self.interval)
	if self.rate == 0.0 {
		self.rate = instantRate
	} else {
		self.rate += self.alpha * (instantRate - self.rate)
	}
}

func (self *EWMA) Rate() float64 {
	return self.rate
}
