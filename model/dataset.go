package model

import (
	"go4ml.xyz/base/tables"
)

/*
Dataset is an abstraction of some source of a data to feed hungry models
*/
type Dataset struct {
	Source     tables.AnyData // It can be tables.Table or lazy stream of mlutil.Struct objects
	Validation tables.AnyData // optional, equal to Source if nil
	Label      string         // name of float32/Tensor field containing label to train
	Test       string         // name of boolean field to select test data
	Features   []string       // patterns of feature names to train model or predict
}
