package model

import (
	"go4ml.xyz/base/fu"
	"go4ml.xyz/base/tables"
	"go4ml.xyz/iokit"
	"go4ml.xyz/zorros"
	"io"
	"path/filepath"
	"reflect"
)

/*
HungryModel is an ML algorithm grows from a data to predict something
Needs to be fattened by Feed method to fit.
*/
type HungryModel interface {
	Feed(Dataset) FatModel
}

/*
Report is an ML training report
*/
type Report struct {
	History     *tables.Table // all iterations history
	TheBest     int           // the best iteration
	Test, Train fu.Struct     // the best iteration metrics
	Score       float64       // the best score
}

/*
Workout is a training iteration abstraction
*/
type Workout interface {
	Iteration() int
	TrainMetrics() MetricsUpdater
	TestMetrics() MetricsUpdater
	Complete(m MemorizeMap, train, test fu.Struct, metricsDone bool) (*Report, bool, error)
	Next() Workout
	Verbose(string)
}

/*
UnifiedTraining is an interface allowing to write any logging/staging backend for ML training
*/
type UnifiedTraining interface {
	// Workout returns the first iteration workout
	Workout() Workout
}

/*
FatModel is fattened model (a training function of model instance bounded to a dataset)
*/
type FatModel func(workout Workout) (*Report, error)

/*
Train a fattened (Fat) model
*/
func (f FatModel) Train(training UnifiedTraining) (*Report, error) {
	w := training.Workout()
	if c, ok := w.(io.Closer); ok {
		defer c.Close()
	}
	return f(w)
}

/*
LuckyTrain trains fattened (Fat) model and trows any occurred errors as a panic
*/
func (f FatModel) LuckyTrain(training UnifiedTraining) *Report {
	m, err := f.Train(training)
	if err != nil {
		panic(zorros.Panic(err))
	}
	return m
}

/*
PredictionModel is a predictor interface
*/
type PredictionModel interface {
	// Features model uses when maps features
	// the same as Features in the training dataset
	Features() []string
	// Column name model adds to result table when maps features.
	// By default it's 'Predicted'
	Predicted() string
	// Returns new table with all original columns except features
	// adding one new column with prediction
	FeaturesMapper(batchSize int) (tables.FeaturesMapper, error)
}

/*
GpuPredictionModel is a prediction interface able to use GPU
*/
type GpuPredictionModel interface {
	PredictionModel
	// Gpu changes context of prediction backend to gpu enabled
	// it's a recommendation only, if GPU is not available or it's impossible to use it
	// the cpu will be used instead
	Gpu(...int) PredictionModel
}

func Path(s string) string {
	if filepath.IsAbs(s) {
		return s
	}
	return iokit.CacheFile(filepath.Join("go-ml", "Models", s))
}

/*
Params is a set of hyper-parameters used by hyper-parameter optimization to generate new model
*/
type Params map[string]float64

/*
Get value of the parameter by name if exists and dflt value otherwise
*/
func (p Params) Get(name string, dflt float64) float64 {
	if v, ok := p[name]; ok {
		return v
	}
	return dflt
}

func (p Params) Apply(m map[string]reflect.Value) {
	for k, v := range p {
		ref, ok := m[k]
		if !ok {
			panic(zorros.Panic(zorros.Errorf("model does not have field `%v`", k)))
		}
		ref.Elem().Set(fu.Convert(reflect.ValueOf(v), false, ref.Type().Elem()))
	}
}
