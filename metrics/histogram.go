package metrics

import (
	"fmt"
	"math"
	"sort"
)

type Sample interface {
	Clear()
	Len() int
	Values() []float64
	Update(value float64)
}

type Histogram struct {
	sample    Sample
	min       float64
	max       float64
	sum       float64
	count     int
	varianceM float64
	varianceS float64
}

func NewHistogram(sample Sample) *Histogram {
	return &Histogram{
		sample:    sample,
		min:       0,
		max:       0,
		sum:       0,
		count:     0,
		varianceM: 0,
		varianceS: 0}
}

/*
  Uses an exponentially decaying sample of 1028 elements, which offers
  a 99.9% confidence level with a 5% margin of error assuming a normal
  distribution, and an alpha factor of 0.015, which heavily biases
  the sample to the past 5 minutes of measurements.
*/
func NewBiasedHistogram() *Histogram {
	return NewHistogram(NewExponentiallyDecayingSample(1028, 0.015))
}

/*
  Uses a uniform sample of 1028 elements, which offers a 99.9%
  confidence level with a 5% margin of error assuming a normal
  distribution.
*/
func NewUnbiasedHistogram() *Histogram {
	return NewHistogram(NewUniformSample(1028))
}

func (self *Histogram) String() string {
	return fmt.Sprintf("Histogram{sum:%.4f count:%d min:%.4f max:%.4f}",
		self.sum, self.count, self.min, self.max)
}

func (self *Histogram) Clear() {
	self.sample.Clear()
	self.min = 0
	self.max = 0
	self.sum = 0
	self.count = 0
	self.varianceM = 0
	self.varianceS = 0
}

func (self *Histogram) Update(value float64) {
	self.count += 1
	self.sum += value
	self.sample.Update(value)
	if self.count == 1 {
		self.min = value
		self.max = value
		self.varianceM = value
	} else {
		if value < self.min {
			self.min = value
		}
		if value > self.max {
			self.max = value
		}
		old_m := self.varianceM
		self.varianceM = old_m + ((value - old_m) / float64(self.count))
		self.varianceS += (value - old_m) * (value - self.varianceM)
	}
}

func (self *Histogram) Count() int {
	return self.count
}

func (self *Histogram) Sum() float64 {
	return self.sum
}

func (self *Histogram) Min() float64 {
	if self.count == 0 {
		return math.NaN()
	}
	return self.min
}

func (self *Histogram) Max() float64 {
	if self.count == 0 {
		return math.NaN()
	}
	return self.max
}

func (self *Histogram) Mean() float64 {
	if self.count > 0 {
		return self.sum / float64(self.count)
	}
	return 0
}

func (self *Histogram) StdDev() float64 {
	if self.count > 0 {
		return math.Sqrt(self.varianceS / float64(self.count-1))
	}
	return 0
}

func (self *Histogram) Variance() float64 {
	if self.count <= 1 {
		return 0
	}
	return self.varianceS / float64(self.count-1)
}

func (self *Histogram) Percentiles(percentiles []float64) []float64 {
	scores := make([]float64, len(percentiles))
	if self.count == 0 {
		return scores
	}

	values := sort.Float64Slice(self.sample.Values())
	sort.Sort(values)
	for i, p := range percentiles {
		pos := p * float64(len(values)+1)
		ipos := int(pos)
		switch {
		case ipos < 1:
			scores[i] = values[0]
		case ipos >= len(values):
			scores[i] = values[len(values)-1]
		default:
			lower := values[ipos-1]
			upper := values[ipos]
			scores[i] = lower + (pos-math.Floor(pos))*(upper-lower)
		}
	}

	return scores
}

func (self *Histogram) Values() []float64 {
	return self.sample.Values()
}
