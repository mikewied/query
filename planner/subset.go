//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package planner

import (
	"github.com/couchbase/query/expression"
)

func SubsetOf(expr1, expr2 expression.Expression) bool {
	v2 := expr2.Value()
	if v2 != nil {
		return v2.Truth()
	}

	s := newSubset(expr1)
	result, _ := expr2.Accept(s)
	return result.(bool)
}

func newSubset(expr expression.Expression) expression.Visitor {
	s, _ := expr.Accept(_SUBSET_FACTORY)
	return s.(expression.Visitor)
}
