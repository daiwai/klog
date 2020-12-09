package workday

import (
	"github.com/stretchr/testify/assert"
	"klog/datetime"
	"testing"
)

func TestAddRanges(t *testing.T) {
	date, _ := datetime.CreateDate(2020, 1, 1)
	time1, _ := datetime.CreateTime(9, 7)
	time2, _ := datetime.CreateTime(12, 59)
	time3, _ := datetime.CreateTime(13, 49)
	time4, _ := datetime.CreateTime(17, 12)
	w := Create(date)
	w.AddRange(time1, time2)
	w.AddRange(time3, time4)
	assert.Equal(t, w.Ranges(), [][]datetime.Time{
		[]datetime.Time{time1, time2},
		[]datetime.Time{time3, time4},
	})
}

func TestAddOpenRange(t *testing.T) {
	date, _ := datetime.CreateDate(2020, 1, 1)
	time1, _ := datetime.CreateTime(9, 7)
	w := Create(date)
	w.AddOpenRange(time1)
	assert.Equal(t, w.Ranges(), [][]datetime.Time{
		[]datetime.Time{time1, nil},
	})
}
