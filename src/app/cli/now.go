package cli

import (
	"errors"
	"fmt"
	. "klog"
	"klog/app"
	"klog/app/cli/lib"
	"klog/service"
	"os"
	"os/signal"
	"syscall"
	gotime "time"
)

type Now struct {
	lib.DiffArgs
	Follow bool `name:"follow" short:"f" help:"Keep shell open and follow changes"`
	lib.WarnArgs
	lib.NoStyleArgs
	lib.InputFilesArgs
}

func (opt *Now) Run(ctx app.Context) error {
	opt.NoStyleArgs.Apply(&ctx)
	h := func() error { return handle(opt, ctx) }
	if opt.Follow {
		return withRepeat(ctx, h)
	}
	return h()
}

func handle(opt *Now, ctx app.Context) error {
	now := ctx.Now()
	records, err := ctx.ReadInputs(opt.File...)
	if err != nil {
		return err
	}
	currentRecords, previousRecords, isToday, err := splitIntoCurrentAndPrevious(now, records)
	if err != nil {
		return err
	}

	INDENT := "          "

	currentTotal, _ := service.HypotheticalTotal(now, currentRecords...)
	currentShouldTotal := service.ShouldTotalSum(currentRecords...)
	currentDiff := service.Diff(currentShouldTotal, currentTotal)
	currentEndTime, _ := NewTimeFromTime(now).Add(NewDuration(0, 0).Minus(currentDiff))

	previousTotal, _ := service.HypotheticalTotal(now, previousRecords...)
	previousShouldTotal := service.ShouldTotalSum(previousRecords...)
	previousDiff := service.Diff(previousShouldTotal, previousTotal)

	grandTotal := currentTotal.Plus(previousTotal)
	grandShouldTotal := NewShouldTotal(0, currentShouldTotal.Plus(previousShouldTotal).InMinutes())
	grandDiff := service.Diff(grandShouldTotal, grandTotal)
	grandEndTime, _ := NewTimeFromTime(now).Add(NewDuration(0, 0).Minus(grandDiff))

	// Headline:
	ctx.Print(INDENT + "   Total")
	if opt.Diff {
		ctx.Print("    Should     Diff   End-Time")
	}
	ctx.Print("\n")

	// Current:
	if isToday {
		ctx.Print("Today    ")
	} else {
		ctx.Print("Yesterday")
	}
	ctx.Print(lib.Pad(9-len(currentTotal.ToString())) + ctx.Serialiser().Duration(currentTotal))
	if opt.Diff {
		ctx.Print(lib.Pad(10-len(currentShouldTotal.ToString())) + ctx.Serialiser().ShouldTotal(currentShouldTotal))
		ctx.Print(lib.Pad(9-len(currentDiff.ToStringWithSign())) + ctx.Serialiser().SignedDuration(currentDiff))
		if currentEndTime != nil {
			ctx.Print(lib.Pad(11-len(currentEndTime.ToString())) + ctx.Serialiser().Time(currentEndTime))
		} else {
			ctx.Print(lib.Pad(11-3) + "???")
		}
	}
	ctx.Print("\n")

	// Previous:
	ctx.Print("Previous")
	ctx.Print(lib.Pad(10-len(previousTotal.ToString())) + ctx.Serialiser().Duration(previousTotal))
	if opt.Diff {
		ctx.Print(lib.Pad(10-len(previousShouldTotal.ToString())) + ctx.Serialiser().ShouldTotal(previousShouldTotal))
		ctx.Print(lib.Pad(9-len(previousDiff.ToStringWithSign())) + ctx.Serialiser().SignedDuration(previousDiff))
	}
	ctx.Print("\n")

	// Line:
	ctx.Print(INDENT + "========")
	if opt.Diff {
		ctx.Print("===================")
	}
	ctx.Print("\n")

	// GrandTotal:
	ctx.Print(INDENT + lib.Pad(7-len(grandTotal.ToString())) + ctx.Serialiser().SignedDuration(grandTotal))
	if opt.Diff {
		ctx.Print(lib.Pad(10-len(grandShouldTotal.ToString())) + ctx.Serialiser().ShouldTotal(grandShouldTotal))
		ctx.Print(lib.Pad(9-len(grandDiff.ToStringWithSign())) + ctx.Serialiser().SignedDuration(grandDiff))
		if grandEndTime != nil {
			ctx.Print(lib.Pad(11-len(grandEndTime.ToString())) + ctx.Serialiser().Time(grandEndTime))
		} else {
			ctx.Print(lib.Pad(11-3) + "???")
		}
	}
	ctx.Print("\n")

	ctx.Print(opt.WarnArgs.ToString(now, records))
	return nil
}

func splitIntoCurrentAndPrevious(now gotime.Time, records []Record) ([]Record, []Record, bool, error) {
	var todaysRecords []Record
	var yesterdaysRecords []Record
	var previousRecords []Record
	today := NewDateFromTime(now)
	yesterday := today.PlusDays(-1)
	for _, r := range records {
		if r.Date().IsEqualTo(today) {
			todaysRecords = append(todaysRecords, r)
		} else if r.Date().IsEqualTo(yesterday) {
			yesterdaysRecords = append(yesterdaysRecords, r)
		} else {
			previousRecords = append(previousRecords, r)
		}
	}
	if len(todaysRecords) > 0 {
		return todaysRecords, append(previousRecords, yesterdaysRecords...), true, nil
	}
	if len(yesterdaysRecords) > 0 {
		return yesterdaysRecords, previousRecords, false, nil
	}
	return nil, nil, false, errors.New("No current record (dated either today or yesterday)")
}

func withRepeat(ctx app.Context, fn func() error) error {
	// Handle ^C gracefully, as it’s the only way to exit
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Exit(0)
		return
	}()

	// Call handler function repetitively
	ctx.Print("\033[2J") // Initial screen clearing
	ticker := gotime.NewTicker(1 * gotime.Second)
	defer ticker.Stop()
	i := 5 // seconds to display help text (how to exit)
	for ; true; <-ticker.C {
		ctx.Print(fmt.Sprintf("\033[H\033[J")) // Cursor reset
		err := fn()
		ctx.Print("\n")
		if i > 0 {
			ctx.Print("Press ^C to exit")
			i--
		}
		if err != nil {
			return err
		}
	}
	return nil
}
