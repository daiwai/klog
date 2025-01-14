package reconciling

import (
	"github.com/jotaen/klog/klog"
	"github.com/jotaen/klog/klog/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestReconcilerAppendingPauseAddsNewEntry(t *testing.T) {
	original := `
2010-04-27
    3:00pm - ?
`
	rs, bs, _ := parser.NewSerialParser().Parse(original)
	reconciler := NewReconcilerAtRecord(klog.Ɀ_Date_(2010, 4, 27))(rs, bs)
	require.NotNil(t, reconciler)
	result, err := reconciler.AppendPause(nil)
	require.Nil(t, err)
	assert.Equal(t, `
2010-04-27
    3:00pm - ?
    -0m
`, result.AllSerialised)
}

func TestReconcilerAppendingPauseWithSummary(t *testing.T) {
	original := `
2010-04-27
    3:00pm - ?
`
	rs, bs, _ := parser.NewSerialParser().Parse(original)
	reconciler := NewReconcilerAtRecord(klog.Ɀ_Date_(2010, 4, 27))(rs, bs)
	require.NotNil(t, reconciler)
	result, err := reconciler.AppendPause(klog.Ɀ_EntrySummary_("Lunch break"))
	require.Nil(t, err)
	assert.Equal(t, `
2010-04-27
    3:00pm - ?
    -0m Lunch break
`, result.AllSerialised)
}

func TestReconcilerAppendingPauseWithMultilineSummary(t *testing.T) {
	original := `
2010-04-27
    3:00pm - ?
`
	rs, bs, _ := parser.NewSerialParser().Parse(original)
	reconciler := NewReconcilerAtRecord(klog.Ɀ_Date_(2010, 4, 27))(rs, bs)
	require.NotNil(t, reconciler)
	result, err := reconciler.AppendPause(klog.Ɀ_EntrySummary_("Lunch", "break"))
	require.Nil(t, err)
	assert.Equal(t, `
2010-04-27
    3:00pm - ?
    -0m Lunch
        break
`, result.AllSerialised)
}

func TestReconcilerAppendPauseWithUTF8Summary(t *testing.T) {
	original := `
2010-04-27
你好！你好吗？
	8:00 - ? 去工作
`
	rs, bs, _ := parser.NewSerialParser().Parse(original)
	reconciler := NewReconcilerAtRecord(klog.Ɀ_Date_(2010, 4, 27))(rs, bs)
	require.NotNil(t, reconciler)
	result, err := reconciler.AppendPause(klog.Ɀ_EntrySummary_("午休"))
	require.Nil(t, err)
	assert.Equal(t, `
2010-04-27
你好！你好吗？
	8:00 - ? 去工作
	-0m 午休
`, result.AllSerialised)
}

func TestReconcilerAppendingPauseFailsIfThereIsNoOpenRange(t *testing.T) {
	original := `
2010-04-27
    3:00 - 4:00
`
	rs, bs, _ := parser.NewSerialParser().Parse(original)
	reconciler := NewReconcilerAtRecord(klog.Ɀ_Date_(2010, 4, 27))(rs, bs)
	require.NotNil(t, reconciler)
	result, err := reconciler.AppendPause(nil)
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestReconcilerExtendingPauseExtendsPause(t *testing.T) {
	original := `
2010-04-27
Foo
    3:00 - ? I desperately need
        a break!
    -30m
`
	rs, bs, _ := parser.NewSerialParser().Parse(original)
	reconciler := NewReconcilerAtRecord(klog.Ɀ_Date_(2010, 4, 27))(rs, bs)
	require.NotNil(t, reconciler)
	result, err := reconciler.ExtendPause(klog.NewDuration(0, -3), nil)
	require.Nil(t, err)
	assert.Equal(t, `
2010-04-27
Foo
    3:00 - ? I desperately need
        a break!
    -33m
`, result.AllSerialised)
}

func TestReconcilerExtendingPauseWithSummary(t *testing.T) {
	original := `
2010-04-27
Foo
    3:00 - ? I desperately need
        a break!
    -1m Lunch break
`
	rs, bs, _ := parser.NewSerialParser().Parse(original)
	reconciler := NewReconcilerAtRecord(klog.Ɀ_Date_(2010, 4, 27))(rs, bs)
	require.NotNil(t, reconciler)
	result, err := reconciler.ExtendPause(klog.NewDuration(0, -3), klog.Ɀ_EntrySummary_("and more break"))
	require.Nil(t, err)
	assert.Equal(t, `
2010-04-27
Foo
    3:00 - ? I desperately need
        a break!
    -4m Lunch break and more break
`, result.AllSerialised)
}

func TestReconcilerExtendingPauseWithSummaryOnNextLine(t *testing.T) {
	original := `
2010-04-27
Foo
    3:00 - ? I desperately need
        a break!
    -1h Lunch break
`
	rs, bs, _ := parser.NewSerialParser().Parse(original)
	reconciler := NewReconcilerAtRecord(klog.Ɀ_Date_(2010, 4, 27))(rs, bs)
	require.NotNil(t, reconciler)
	result, err := reconciler.ExtendPause(klog.NewDuration(-1, 0), klog.Ɀ_EntrySummary_("", "and more break"))
	require.Nil(t, err)
	assert.Equal(t, `
2010-04-27
Foo
    3:00 - ? I desperately need
        a break!
    -2h Lunch break
        and more break
`, result.AllSerialised)
}

func TestReconcilerExtendingPauseWithMultilineSummary(t *testing.T) {
	original := `
2010-04-27
Foo
    3:00 - ? I desperately need
        a break!
    -1h Lunch
        break
`
	rs, bs, _ := parser.NewSerialParser().Parse(original)
	reconciler := NewReconcilerAtRecord(klog.Ɀ_Date_(2010, 4, 27))(rs, bs)
	require.NotNil(t, reconciler)
	result, err := reconciler.ExtendPause(klog.NewDuration(-1, 0), klog.Ɀ_EntrySummary_("and more break"))
	require.Nil(t, err)
	assert.Equal(t, `
2010-04-27
Foo
    3:00 - ? I desperately need
        a break!
    -2h Lunch
        break and more break
`, result.AllSerialised)
}

func TestReconcilerExtendingPauseExtendsLastPause(t *testing.T) {
	original := `
2010-04-27
Foo
    3:00 - ? I desperately need
        a break!
    -30m
    -30m
`
	rs, bs, _ := parser.NewSerialParser().Parse(original)
	reconciler := NewReconcilerAtRecord(klog.Ɀ_Date_(2010, 4, 27))(rs, bs)
	require.NotNil(t, reconciler)
	result, err := reconciler.ExtendPause(klog.NewDuration(-2, -51), nil)
	require.Nil(t, err)
	assert.Equal(t, `
2010-04-27
Foo
    3:00 - ? I desperately need
        a break!
    -30m
    -3h21m
`, result.AllSerialised)
}

func TestReconcilerExtendingPauseOfZeroIsNoop(t *testing.T) {
	original := `
2010-04-27
Foo
    3:00 - ? I desperately need
        a break!
    -0m
`
	rs, bs, _ := parser.NewSerialParser().Parse(original)
	reconciler := NewReconcilerAtRecord(klog.Ɀ_Date_(2010, 4, 27))(rs, bs)
	require.NotNil(t, reconciler)
	result, err := reconciler.ExtendPause(klog.NewDuration(0, 0), nil)
	require.Nil(t, err)
	assert.Equal(t, `
2010-04-27
Foo
    3:00 - ? I desperately need
        a break!
    -0m
`, result.AllSerialised)
}

func TestReconcilerExtendingPauseFailsIfThereIsNoOpenRange(t *testing.T) {
	original := `
2010-04-27
    3:00 - 4:00
    -30m
`
	rs, bs, _ := parser.NewSerialParser().Parse(original)
	reconciler := NewReconcilerAtRecord(klog.Ɀ_Date_(2010, 4, 27))(rs, bs)
	require.NotNil(t, reconciler)
	result, err := reconciler.ExtendPause(klog.NewDuration(2, 0), nil)
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestReconcilerDoesNotExtendNonNegativeDurations(t *testing.T) {
	original := `
2010-04-27
    3:00 - ?
    30m
`
	rs, bs, _ := parser.NewSerialParser().Parse(original)
	reconciler := NewReconcilerAtRecord(klog.Ɀ_Date_(2010, 4, 27))(rs, bs)
	require.NotNil(t, reconciler)
	result, err := reconciler.ExtendPause(klog.NewDuration(0, -10), nil)
	require.Error(t, err)
	assert.Nil(t, result)
}
