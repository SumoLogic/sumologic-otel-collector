package main

import "github.com/mikefarah/yq/v4/pkg/yqlib"

func init() {
	// It seems that this isn't necessarily called by yqlib all the time for
	// some reason. This ensures that the ExpressionParser global is not nil.
	yqlib.InitExpressionParser()
}
