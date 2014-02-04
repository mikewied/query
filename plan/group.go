//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package plan

import (
	_ "fmt"

	"github.com/couchbaselabs/query/algebra"
)

// Grouping of input data.
type InitialGroup struct {
	keys algebra.ExpressionList
	aggs algebra.ExpressionList
}

// Grouping of groups. Recursable.
type IntermediateGroup struct {
	keys algebra.ExpressionList
	aggs algebra.ExpressionList
}

// Compute DistinctCount() and Avg().
type FinalGroup struct {
	keys algebra.ExpressionList
	aggs algebra.ExpressionList
}

func NewInitialGroup(keys, aggs algebra.ExpressionList) *InitialGroup {
	return &InitialGroup{keys, aggs}
}

func (this *InitialGroup) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitInitialGroup(this)
}

func NewIntermediateGroup(keys, aggs algebra.ExpressionList) *IntermediateGroup {
	return &IntermediateGroup{keys, aggs}
}

func (this *IntermediateGroup) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitIntermediateGroup(this)
}

func NewFinalGroup(keys, aggs algebra.ExpressionList) *FinalGroup {
	return &FinalGroup{keys, aggs}
}

func (this *FinalGroup) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitFinalGroup(this)
}