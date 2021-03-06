//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package algebra

import (
	"github.com/couchbase/query/datastore"
	"github.com/couchbase/query/errors"
	"github.com/couchbase/query/expression"
	"github.com/couchbase/query/value"
)

/*
This represents the select statement. Type Select is a
struct that contains fields mapping to each clause in
a select statement. The subresult field maps to the
intermediate result interface for the select clause.
The order field maps to the order by clause, the offset
is an expression that maps to the offset clause and
similarly limit is an expression that maps to the limit
clause.
*/
type Select struct {
	statementBase

	subresult Subresult             `json:"subresult"`
	order     *Order                `json:"order"`
	offset    expression.Expression `json:"offset"`
	limit     expression.Expression `json:"limit"`
}

/*
The function NewSelect returns a pointer to the Select struct
by assigning the input attributes to the fields of the struct.
*/
func NewSelect(subresult Subresult, order *Order, offset, limit expression.Expression) *Select {
	rv := &Select{
		subresult: subresult,
		order:     order,
		offset:    offset,
		limit:     limit,
	}

	rv.stmt = rv
	return rv
}

/*
Visitor pattern.
*/
func (this *Select) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitSelect(this)
}

/*
This method returns the shape of this statement.
*/
func (this *Select) Signature() value.Value {
	return this.subresult.Signature()
}

/*
This method calls FormalizeSubquery to qualify all the children
of the query, and returns an error if any.
*/
func (this *Select) Formalize() (err error) {
	return this.FormalizeSubquery(expression.NewFormalizer())
}

/*
This method maps all the constituent clauses, namely the subresult,
order, limit and offset within a Select statement.
*/
func (this *Select) MapExpressions(mapper expression.Mapper) (err error) {
	err = this.subresult.MapExpressions(mapper)
	if err != nil {
		return
	}

	if this.order != nil {
		err = this.order.MapExpressions(mapper)
	}

	if this.limit != nil {
		this.limit, err = mapper.Map(this.limit)
		if err != nil {
			return
		}
	}

	if this.offset != nil {
		this.offset, err = mapper.Map(this.offset)
	}

	return
}

/*
   Returns all contained Expressions.
*/
func (this *Select) Expressions() expression.Expressions {
	exprs := this.subresult.Expressions()

	if this.order != nil {
		exprs = append(exprs, this.order.Expressions()...)
	}

	if this.limit != nil {
		exprs = append(exprs, this.limit)
	}

	if this.offset != nil {
		exprs = append(exprs, this.offset)
	}

	return exprs
}

/*
Returns all required privileges.
*/
func (this *Select) Privileges() (datastore.Privileges, errors.Error) {
	privs, err := this.subresult.Privileges()
	if err != nil {
		return nil, err
	}

	exprs := make(expression.Expressions, 0, 16)

	if this.order != nil {
		exprs = append(exprs, this.order.Expressions()...)
	}

	if this.limit != nil {
		exprs = append(exprs, this.limit)
	}

	if this.offset != nil {
		exprs = append(exprs, this.offset)
	}

	subprivs, err := subqueryPrivileges(exprs)
	if err != nil {
		return nil, err
	}

	privs.Add(subprivs)
	return privs, nil
}

/*
   Representation as a N1QL string.
*/
func (this *Select) String() string {
	s := this.subresult.String()

	if this.order != nil {
		s += " " + this.order.String()
	}

	if this.limit != nil {
		s += " limit " + this.limit.String()
	}

	if this.offset != nil {
		s += " offset " + this.offset.String()
	}

	return s
}

/*
This method qualifies identifiers for all the constituent clauses,
namely the subresult, order, limit and offset within a subquery.
For the subresult of the subquery, call Formalize, for the order
by clause call MapExpressions, for limit and offset call Accept.
*/
func (this *Select) FormalizeSubquery(parent *expression.Formalizer) (err error) {
	formalizer, err := this.subresult.Formalize(parent)
	if err != nil {
		return err
	}

	if this.order != nil && formalizer.Keyspace != "" {
		err = this.order.MapExpressions(formalizer)
		if err != nil {
			return
		}
	}

	if this.limit != nil {
		_, err = this.limit.Accept(parent)
		if err != nil {
			return
		}
	}

	if this.offset != nil {
		_, err = this.offset.Accept(parent)
		if err != nil {
			return
		}
	}

	return
}

/*
Return the subresult of the select statement.
*/
func (this *Select) Subresult() Subresult {
	return this.subresult
}

/*
Return the order by clause in the select statement.
*/
func (this *Select) Order() *Order {
	return this.order
}

/*
Returns the offset expression in the select clause.
*/
func (this *Select) Offset() expression.Expression {
	return this.offset
}

/*
Returns the limit expression in the select clause.
*/
func (this *Select) Limit() expression.Expression {
	return this.limit
}

/*
This method sets the limit expression for the select
statement.
*/
func (this *Select) SetLimit(limit expression.Expression) {
	this.limit = limit
}

/*
The Subresult interface represents the intermediate result of a
select statement. It inherits from Node.
*/
type Subresult interface {
	/*
	   Inherts Node. The Node interface represents a node in
	   the algebra tree (AST).
	*/
	Node

	/*
	   The shape of this statement's return values.
	*/
	Signature() value.Value

	/*
	   Fully qualify all identifiers in this statement.
	*/
	Formalize(parent *expression.Formalizer) (formalizer *expression.Formalizer, err error)

	/*
	   Apply a Mapper to all the expressions in this statement
	*/
	MapExpressions(mapper expression.Mapper) error

	/*
	   Returns all contained Expressions.
	*/
	Expressions() expression.Expressions

	/*
	   Returns all required privileges.
	*/
	Privileges() (datastore.Privileges, errors.Error)

	/*
	   Representation as a N1QL string.
	*/
	String() string

	/*
	   Checks if correlated subquery.
	*/
	IsCorrelated() bool
}
