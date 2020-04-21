package model

import (
	"fmt"
	"go-ml.dev/pkg/base/fu"
	"go-ml.dev/pkg/base/fu/lazy"
	"go-ml.dev/pkg/base/tables"
	"go-ml.dev/pkg/iokit"
	"go-ml.dev/pkg/zorros"
	"go-ml.dev/pkg/zorros/zlog"
	"io"
	"reflect"
)

/*
Training is the default implementation of unified training interface
*/
type Training struct {
	Iterations   int          // maximum iterations
	Metrics      Metrics      // evaluating metrics
	Score        Score        // score function
	ScoreHistory int          // possible count of forehead training with lower score
	ModelFile    iokit.Output // file to store final model
	Verbose      interface{}  // print function func(string)
}

type training struct {
	Training
	stash *ModelStash
	done  bool
}

type workout struct {
	iteration int
	training  *training
	perflog   [][2]fu.Struct
	scorlog   []float64
}

const DefaultScoreHistory = 3

func (t Training) Workout() Workout {
	x := &training{
		Training: t,
		stash:    NewStash(fu.Fnzi(t.ScoreHistory, DefaultScoreHistory), "model-treaining-*.zip"),
	}
	return &workout{iteration: 0, training: x}
}

func (w *workout) Close() error {
	return w.training.stash.Close()
}

func (w *workout) Iteration() int {
	return w.iteration
}

func (w *workout) TrainMetrics() MetricsUpdater {
	return w.training.Metrics.New(w.iteration, TrainSubset)
}

func (w *workout) TestMetrics() MetricsUpdater {
	return w.training.Metrics.New(w.iteration, TestSubset)
}

func (w *workout) report(j int) (report *Report, err error) {
	report = &Report{}
	histlen := fu.Fnzi(w.training.ScoreHistory, DefaultScoreHistory)
	if len(w.perflog) > 0 {
		report.History = tables.Lazy(lazy.Flatn(w.perflog)).LuckyCollect()
		if j == 0 {
			l := fu.Mini(len(w.scorlog), histlen)
			lj := len(w.scorlog) - l
			j = fu.Indmaxd(w.scorlog[lj:]) + lj
		}
		report.TheBest = j
		report.Train = w.perflog[j][0]
		report.Test = w.perflog[j][1]
		report.Score = w.scorlog[j]
		if w.training.ModelFile != nil {
			rd, e := w.training.stash.Reader(j)
			if e != nil {
				err = zorros.Trace(e)
				return
			}
			wh, e := w.training.ModelFile.Create()
			if e != nil {
				err = zorros.Trace(e)
				return
			}
			defer wh.End()
			_, e = io.Copy(wh, rd)
			if e != nil {
				err = zorros.Trace(e)
				return
			}
			if e = wh.Commit(); e != nil {
				err = zorros.Trace(e)
				return
			}
		}
	} else {
		report.History = tables.NewEmpty(w.training.Metrics.Names(), nil)
	}
	return
}

func (w *workout) Complete(m MemorizeMap, train, test fu.Struct, metricsDone bool) (report *Report, done bool, err error) {
	histlen := fu.Fnzi(w.training.ScoreHistory, DefaultScoreHistory)
	maxiter := fu.Maxi(w.training.Iterations, 1)
	score := w.training.Score(train, test)
	w.scorlog = append(w.scorlog, score)
	w.perflog = append(w.perflog, [2]fu.Struct{train, test})
	if w.training.ModelFile != nil {
		o, e := w.training.stash.Output(w.iteration)
		if e != nil {
			err = zorros.Wrapf(e, "failed to create stash for model: %v", e.Error())
			return
		}
		if err = Memorize(o, m); err != nil {
			return
		}
	}
	if metricsDone {
		w.training.done = true
		done = true
		report, err = w.report(w.iteration)
	} else if w.iteration == maxiter-1 || (w.iteration > histlen && fu.Indmaxd(w.scorlog[len(w.scorlog)-histlen:]) == 0) {
		w.training.done = true
		done = true
		report, err = w.report(0)
	}
	if w.training.Verbose != nil {
		w.Verbose(fmt.Sprintf(
			"[%3d] loss: %.5f/%.5f, error: %.5f/%.5f, score: %.5f",
			w.Iteration(), Loss(train), Loss(test), Error(train), Error(test), score))
	}
	return
}

func (w *workout) Verbose(s string) {
	if w.training.Verbose != nil {
		vf := reflect.ValueOf(w.training.Verbose)
		vf.Call([]reflect.Value{reflect.ValueOf(s)})
	}
}

func (w *workout) Next() Workout {
	if w.training == nil {
		//panic(zorros.Panic(zorros.Errorf("training is done")))
		zlog.Warning("training is already done")
		return nil
	}
	return &workout{
		iteration: w.iteration + 1,
		training:  w.training,
		scorlog:   w.scorlog,
		perflog:   w.perflog,
	}
}
