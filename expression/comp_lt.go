//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package expression

import (
	"github.com/couchbase/query/value"
)

/*
Comparison terms allow for comparing two expressions.
This represents the less than comparison
operation. Type LT is a struct that implements
BinaryFunctionBase.
*/
type LT struct {
	BinaryFunctionBase
}

/*
The function NewLT calls NewBinaryFunctionBase
to define less than comparison expression
with input operand expressions first and second,
as input.
*/
func NewLT(first, second Expression) Function {
	rv := &LT{
		*NewBinaryFunctionBase("lt", first, second),
	}

	rv.expr = rv
	return rv
}

/*
It calls the VisitLT method by passing in the receiver to
and returns the interface. It is a visitor pattern.
*/
func (this *LT) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitLT(this)
}

/*
It returns a value type BOOLEAN.
*/
func (this *LT) Type() value.Type { return value.BOOLEAN }

/*
Calls the Eval method for Binary functions and passes in the
receiver, current item and current context.
*/
func (this *LT) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.BinaryEval(this, item, context)
}

func (this *LT) Apply(context Context, first, second value.Value) (value.Value, error) {
	cmp := first.Compare(second)
	switch actual := cmp.Actual().(type) {
	case float64:
		return value.NewValue(actual < 0), nil
	}

	return cmp, nil
}

/*
The constructor returns a NewLT with the operands
cast to a Function as the FunctionConstructor.
*/
func (this *LT) Constructor() FunctionConstructor {
	return func(operands ...Expression) Function {
		return NewLT(operands[0], operands[1])
	}
}
