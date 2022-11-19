package cli

import (
	"github.com/jotaen/klog/klog"
	"github.com/jotaen/klog/klog/app"
	"github.com/jotaen/klog/klog/app/cli/lib"
	"github.com/jotaen/klog/klog/parser/reconciling"
	"strings"
	gotime "time"
)

type Pause struct {
	Summary klog.EntrySummary `name:"summary" short:"s" placeholder:"TEXT" help:"Summary text for the pause entry"`
	lib.OutputFileArgs
	lib.WarnArgs
}

func (opt *Pause) Help() string {
	return `This doesn’t actually stop the open-ended time range.
Instead, it adds/extends an entry underneath the open-ended time range that contains the duration of the pause.
The command is blocking, and it keeps updating the pause entry until the process is exited.
(The file will be written into once per minute.)
`
}

func (opt *Pause) Run(ctx app.Context) app.Error {
	// We don’t rely on the accumulated counter, because then it might
	// also accumulate imprecisions over time. Therefore, we always base the
	// increment off the initial start time.
	//
	// Upon initial invocation, it performs one dry-run. This is for detecting – and,
	// more importantly: displaying – errors right away; like malicious syntax,
	// or if there is no open-ended time range.
	start := gotime.Now()
	minsProcessed := 0
	isDryRun := true
	return lib.WithRepeat(ctx.Print, 500*gotime.Millisecond, func(counter int64) app.Error {
		dots := strings.Repeat(".", int(counter%4))
		diffMins := int(-1 * gotime.Time.Sub(start, gotime.Now()).Minutes())
		ctx.Print("Pausing since " +
			klog.NewTimeFromGo(start).ToString() +
			" (" +
			ctx.Serialiser().Duration(klog.NewDuration(0, diffMins)) +
			")" + dots + "\n")
		if counter < 14 {
			// Display exit hint for a couple of seconds.
			ctx.Print("\n")
			ctx.Print("Press ^C to stop\n")
		}
		increment := diffMins - minsProcessed
		if !isDryRun && increment <= 0 {
			return nil
		}
		minsProcessed += increment
		today := klog.NewDateFromGo(gotime.Now())
		_, err := ctx.ReconcileFile(
			!isDryRun,
			opt.OutputFileArgs.File,
			[]reconciling.Creator{
				reconciling.NewReconcilerAtRecord(today),
				reconciling.NewReconcilerAtRecord(today.PlusDays(-1)),
			},

			func(reconciler *reconciling.Reconciler) (*reconciling.Result, error) {
				return reconciler.PauseOpenRange(klog.NewDuration(0, -1*increment), opt.Summary)
			},
		)
		isDryRun = false
		return err
	})
}
