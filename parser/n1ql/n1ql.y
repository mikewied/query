%{
package n1ql

import "github.com/couchbaselabs/clog"
import "github.com/couchbaselabs/query/algebra"
import "github.com/couchbaselabs/query/catalog"
import "github.com/couchbaselabs/query/expression"
import "github.com/couchbaselabs/query/value"

func logDebugGrammar(format string, v ...interface{}) {
    clog.To("PARSER", format, v...)
}
%}

%union {
s string
n int
f float64
b bool

expr             expression.Expression
exprs            expression.Expressions
whenTerm         *expression.WhenTerm
whenTerms        expression.WhenTerms
binding          *expression.Binding
bindings         expression.Bindings

explain          *algebra.Explain

statement        algebra.Statement

fullselect       *algebra.Select
subresult        algebra.Subresult
subselect        *algebra.Subselect
fromTerm         algebra.FromTerm
bucketTerm       *algebra.BucketTerm
path             expression.Path
group            *algebra.Group
resultTerm       *algebra.ResultTerm
resultTerms      algebra.ResultTerms
projection       *algebra.Projection
sortTerm         *algebra.SortTerm
sortTerms        algebra.SortTerms

bucketRef        *algebra.BucketRef

set              *algebra.Set
unset            *algebra.Unset
setTerm          *algebra.SetTerm
setTerms         algebra.SetTerms
unsetTerm        *algebra.UnsetTerm
unsetTerms       algebra.UnsetTerms
updateFor        *algebra.UpdateFor
mergeActions     *algebra.MergeActions
mergeUpdate      *algebra.MergeUpdate
mergeDelete      *algebra.MergeDelete
mergeInsert      *algebra.MergeInsert

createIndex      *algebra.CreateIndex
dropIndex        *algebra.DropIndex
alterIndex       *algebra.AlterIndex
indexType        catalog.IndexType
}

%token ALL
%token ALTER
%token AND
%token ANY
%token ARRAY
%token AS
%token ASC
%token BETWEEN
%token BUCKET
%token BY
%token CASE
%token CAST
%token CLUSTER
%token COLLATE
%token CREATE
%token DATABASE
%token DATASET
%token DATATAG
%token DELETE
%token DESC
%token DISTINCT
%token DROP
%token EACH
%token EXCEPT
%token EXISTS
%token ELSE
%token END
%token EVERY
%token EXISTS
%token EXPLAIN
%token FALSE
%token FIRST
%token FOR
%token FROM
%token GROUP
%token HAVING
%token IF
%token IN
%token INDEX
%token INLINE
%token INNER
%token INSERT
%token INTERSECT
%token INTO
%token IS
%token JOIN
%token KEY
%token KEYS
%token LEFT
%token LET
%token LETTING
%token LIKE
%token LIMIT
%token MATCHED
%token MERGE
%token MISSING
%token NEST
%token NOT
%token NULL
%token OFFSET
%token ON
%token OR
%token ORDER
%token OUTER
%token OVER
%token PARTITION
%token PATH
%token POOL
%token PRIMARY
%token RAW
%token REALM
%token RENAME
%token RETURNING
%token RIGHT
%token SATISFIES
%token SET
%token SOME
%token SELECT
%token THEN
%token TO
%token TRUE
%token UNDER
%token UNION
%token UNIQUE
%token UNNEST
%token UNSET
%token UPDATE
%token UPSERT
%token USING
%token VALUED
%token VALUES
%token VIEW
%token WHERE
%token WHEN
%token WITH
%token XOR

%token INT NUMBER IDENTIFIER STRING
%token LPAREN RPAREN
%token LBRACE RBRACE LBRACKET RBRACKET
%token COMMA COLON

/* Precedence: lowest to highest */
%left           UNION EXCEPT
%left           INTERSECT
%left           JOIN NEST UNNEST INNER LEFT
%left           OR
%left           AND
%right          NOT
%nonassoc       EQ DEQ NE
%nonassoc       LT GT LE GE
%nonassoc       LIKE
%nonassoc       BETWEEN
%nonassoc       IN
%nonassoc       EXISTS
%nonassoc       IS                              /* IS NULL, IS MISSING, IS VALUED, IS NOT NULL, etc. */
%left           CONCAT
%left           PLUS MINUS
%left           STAR DIV MOD

/* Unary operators */
%right          UMINUS
%left           DOT LBRACKET RBRACKET

/* Override precedence */
%left           LPAREN RPAREN

/* Types */
%type <s>                STRING
%type <s>                IDENTIFIER
%type <f>                NUMBER
%type <n>                INT
%type <expr>             literal object array
%type <binding>          member
%type <bindings>         members opt_members

%type <expr>             expr c_expr b_expr
%type <exprs>            exprs opt_exprs
%type <binding>          binding
%type <bindings>         bindings

%type <s>                alias as_alias opt_as_alias variable

%type <expr>             case_expr simple_or_searched_case simple_case searched_case opt_else
%type <whenTerms>        when_thens

%type <expr>             collection_expr collection_cond collection_xform
%type <binding>          coll_binding
%type <bindings>         coll_bindings
%type <expr>             satisfies
%type <expr>             opt_when

%type <expr>             function_expr
%type <s>                function_name

%type <expr>             paren_or_subquery_expr paren_or_subquery

%type <explain>          explain
%type <fullselect>       fullselect
%type <subresult>        subselects
%type <subselect>        subselect
%type <subselect>        select_from
%type <subselect>        from_select
%type <fromTerm>         from_term from opt_from
%type <bucketTerm>       bucket_term
%type <b>                opt_join_type
%type <path>             path opt_subpath
%type <s>                pool_name bucket_name
%type <expr>             keys opt_keys
%type <bindings>         opt_let let
%type <expr>             opt_where where
%type <group>            opt_group group
%type <bindings>         opt_letting letting
%type <expr>             opt_having having
%type <resultTerm>       project
%type <resultTerms>      projects
%type <projection>       projection select_clause
%type <sortTerm>         sort_term
%type <sortTerms>        sort_terms order_by opt_order_by
%type <expr>             limit opt_limit
%type <expr>             offset opt_offset
%type <b>                dir opt_dir

%type <statement>        stmt select_stmt dml_stmt ddl_stmt
%type <statement>        insert upsert delete update merge
%type <statement>        index_stmt create_index drop_index alter_index

%type <bucketRef>        bucket_ref
%type <exprs>            values
%type <expr>             key opt_key
%type <projection>       returns returning opt_returning
%type <binding>          update_binding
%type <bindings>         update_bindings
%type <expr>             path_expr
%type <set>              set
%type <setTerm>          set_term
%type <setTerms>         set_terms
%type <unset>            unset
%type <unsetTerm>        unset_term
%type <unsetTerms>       unset_terms
%type <updateFor>        update_for opt_update_for
%type <binding>          update_binding
%type <bindings>         update_bindings
%type <mergeActions>     merge_actions opt_merge_delete_insert
%type <mergeUpdate>      merge_update
%type <mergeDelete>      merge_delete
%type <mergeInsert>      merge_insert opt_merge_insert

%type <s>                index_name
%type <bucketRef>        named_bucket_ref
%type <expr>             index_partition
%type <indexType>        index_using
%type <s>                rename

%start input

%%

input:
explain
|
stmt
;

explain:
EXPLAIN stmt
{
  $$ = algebra.NewExplain($2)
}
;

stmt:
select_stmt
|
dml_stmt
|
ddl_stmt
;

select_stmt:
fullselect
{
  $$ = $1
}
;

dml_stmt:
insert
|
upsert
|
delete
|
update
|
merge
;

ddl_stmt:
index_stmt
;

index_stmt:
create_index
|
drop_index
|
alter_index
;

fullselect:
subselects opt_order_by opt_limit opt_offset
{
  $$ = algebra.NewSelect($1, $2, $4, $3) /* OFFSET precedes LIMIT */
}
;

subselects:
subselect
{
  $$ = $1
}
|
subselects UNION subselect
{
  $$ = algebra.NewUnion($1, $3)
}
|
subselects UNION ALL subselect
{
  $$ = algebra.NewUnionAll($1, $4)
}
;

subselect:
from_select
|
select_from
;

from_select:
from opt_let opt_where opt_group select_clause
{
  $$ = algebra.NewSubselect($1, $2, $3, $4, $5)
}
;

select_from:
select_clause opt_from opt_let opt_where opt_group
{
  $$ = algebra.NewSubselect($2, $3, $4, $5, $1)
}
;


/*************************************************
 *
 * SELECT clause
 *
 *************************************************/

select_clause:
SELECT
projection
{
  $$ = $2
}
;

projection:
projects
{
  $$ = algebra.NewProjection(false, $1)
}
|
DISTINCT projects
{
  $$ = algebra.NewProjection(true, $2)
}
|
ALL projects
{
  $$ = algebra.NewProjection(false, $2)
}
|
RAW expr
{
  $$ = algebra.NewRawProjection(false, $2)
}
|
DISTINCT RAW expr
{
  $$ = algebra.NewRawProjection(true, $3)
}
;

projects:
project
{
  $$ = algebra.ResultTerms{$1}
}
|
projects COMMA project
{
  $$ = append($1, $3)
}
;

project:
STAR
{
  $$ = algebra.NewResultTerm(nil, true, "")
}
|
expr DOT STAR
{
  $$ = algebra.NewResultTerm($1, true, "")
}
|
expr opt_as_alias
{
  $$ = algebra.NewResultTerm($1, false, $2)
}
;

opt_as_alias:
/* empty */
{
  $$ = ""
}
|
as_alias
;

as_alias:
alias
|
AS alias
{
  $$ = $2
}
;

alias:
IDENTIFIER
;


/*************************************************
 *
 * FROM clause
 *
 *************************************************/

opt_from:
/* empty */
{
  $$ = nil
}
|
from
;

from:
FROM from_term
{
  $$ = $2
}
;

from_term:
bucket_term
{
  $$ = $1
}
|
from_term opt_join_type JOIN bucket_term
{
  $$ = algebra.NewJoin($1, $2, $4)
}
|
from_term opt_join_type NEST bucket_term
{
  $$ = algebra.NewNest($1, $2, $4)
}
|
from_term opt_join_type UNNEST path opt_as_alias
{
  $$ = algebra.NewUnnest($1, $2, $4, $5)
}
;

bucket_term:
pool_name COLON bucket_name opt_subpath opt_as_alias opt_keys
{
  $$ = algebra.NewBucketTerm($1, $3, $4, $5, $6)
}
|
bucket_name opt_subpath opt_as_alias opt_keys
{
  $$ = algebra.NewBucketTerm("", $1, $2, $3, $4)
}
;

pool_name:
IDENTIFIER
;

bucket_name:
IDENTIFIER
;

opt_subpath:
/* empty */
{
  $$ = nil
}
|
DOT path
{
  $$ = $2
}
;

opt_keys:
/* empty */
{
  $$ = nil
}
|
keys
;

keys:
KEYS expr
{
  $$ = $2
}
;

opt_join_type:
/* empty */
{
  $$ = false
}
|
INNER
{
  $$ = false
}
|
LEFT opt_outer
{
  $$ = true
}
;

opt_outer:
/* empty */
|
OUTER
;


/*************************************************
 *
 * LET clause
 *
 *************************************************/

opt_let:
/* empty */
{
  $$ = nil
}
|
let
;

let:
LET bindings
{
  $$ = $2
}
;

bindings:
binding
{
  $$ = expression.Bindings{$1}
}
|
bindings COMMA binding
{
  $$ = append($1, $3)
}
;

binding:
alias EQ expr
{
  $$ = expression.NewBinding($1, $3)
}
;


/*************************************************
 *
 * WHERE clause
 *
 *************************************************/

opt_where:
/* empty */
{
  $$ = nil
}
|
where
;

where:
WHERE expr
{
  $$ = $2
}
;


/*************************************************
 *
 * GROUP BY clause
 *
 *************************************************/

opt_group:
/* empty */
{
  $$ = nil
}
|
group
;

group:
GROUP BY exprs opt_letting opt_having
{
  $$ = algebra.NewGroup($3, $4, $5)
}
;

exprs:
expr
{
  $$ = expression.Expressions{$1}
}
|
exprs COMMA expr
{
  $$ = append($1, $3)
}
;

opt_letting:
/* empty */
{
  $$ = nil
}
|
letting
;

letting:
LETTING bindings
{
  $$ = $2
}
;

opt_having:
/* empty */
{
  $$ = nil
}
|
having
;

having:
HAVING expr
{
  $$ = $2
}
;


/*************************************************
 *
 * ORDER BY clause
 *
 *************************************************/

opt_order_by:
/* empty */
{
  $$ = nil
}
|
order_by
;

order_by:
ORDER BY sort_terms
{
  $$ = $3
}
;

sort_terms:
sort_term
{
  $$ = algebra.SortTerms{$1}
}
|
sort_terms COMMA sort_term
{
  $$ = append($1, $3)
}
;

sort_term:
expr opt_dir
{
  $$ = algebra.NewSortTerm($1, $2)
}
;

opt_dir:
/* empty */
{
  $$ = false
}
|
dir
;

dir:
ASC
{
  $$ = false
}
|
DESC
{
  $$ = true
}
;


/*************************************************
 *
 * LIMIT clause
 *
 *************************************************/

opt_limit:
/* empty */
{
  $$ = nil
}
|
limit
;

limit:
LIMIT expr
{
  $$ = $2
}
;


/*************************************************
 *
 * OFFSET clause
 *
 *************************************************/

opt_offset:
/* empty */
{
  $$ = nil
}
|
offset
;

offset:
OFFSET expr
{
  $$ = $2
}
;


/*************************************************
 *
 * INSERT
 *
 *************************************************/

insert:
INSERT INTO bucket_ref opt_key values opt_returning
{
  $$ = algebra.NewInsertValues($3, $4, $5, $6)
}
|
INSERT INTO bucket_ref opt_key fullselect opt_returning
{
  $$ = algebra.NewInsertSelect($3, $4, $5, $6)
}
;

bucket_ref:
pool_name COLON bucket_name opt_as_alias
{
  $$ = algebra.NewBucketRef($1, $3, $4)
}
|
bucket_name opt_as_alias
{
  $$ = algebra.NewBucketRef("", $1, $2)
}
;

opt_key:
/* empty */
{
  $$ = nil
}
|
key
;

key:
KEY expr
{
  $$ = $2
}
;

values:
VALUES exprs
{
  $$ = $2
}
;

opt_returning:
/* empty */
{
  $$ = nil
}
|
returning
;

returning:
RETURNING returns
{
  $$ = $2
}
;

returns:
projects
{
  $$ = algebra.NewProjection(false, $1)
}
|
RAW expr
{
  $$ = algebra.NewRawProjection(false, $2)
}
;


/*************************************************
 *
 * UPSERT
 *
 *************************************************/

upsert:
UPSERT INTO bucket_ref key values opt_returning
{
  $$ = algebra.NewUpsertValues($3, $4, $5, $6)
}
|
UPSERT INTO bucket_ref key fullselect opt_returning
{
  $$ = algebra.NewUpsertSelect($3, $4, $5, $6)
}
;


/*************************************************
 *
 * DELETE
 *
 *************************************************/

delete:
DELETE FROM bucket_ref opt_keys opt_where opt_limit opt_returning
{
  $$ = algebra.NewDelete($3, $4, $5, $6, $7)
}
;


/*************************************************
 *
 * UPDATE
 *
 *************************************************/

update:
UPDATE bucket_ref opt_keys set unset opt_where opt_limit opt_returning
{
  $$ = algebra.NewUpdate($2, $3, $4, $5, $6, $7, $8)
}
|
UPDATE bucket_ref opt_keys set opt_where opt_limit opt_returning
{
  $$ = algebra.NewUpdate($2, $3, $4, nil, $5, $6, $7)
}
|
UPDATE bucket_ref opt_keys unset opt_where opt_limit opt_returning
{
  $$ = algebra.NewUpdate($2, $3, nil, $4, $5, $6, $7)
}
;

set:
SET set_terms
{
  $$ = algebra.NewSet($2)
}
;

set_terms:
set_term
{
  $$ = algebra.SetTerms{$1}
}
|
set_terms COMMA set_term
{
  $$ = append($1, $3)
}
;

set_term:
path EQ expr opt_update_for
{
  $$ = algebra.NewSetTerm($1, $3, $4)
}
;

opt_update_for:
/* empty */
{
  $$ = nil
}
|
update_for
;

update_for:
FOR update_bindings opt_when END
{
  $$ = algebra.NewUpdateFor($2, $3)
}
;

update_bindings:
update_binding
{
  $$ = expression.Bindings{$1}
}
|
update_bindings COMMA update_binding
{
  $$ = append($1, $3)
}
;

update_binding:
variable IN path_expr
{
  $$ = expression.NewBinding($1, $3)
}
;

variable:
IDENTIFIER
;

path_expr:
path
{
  $$ = $1
}
;

opt_when:
/* empty */
{
  $$ = nil
}
|
WHEN expr
{
  $$ = $2
}
;

unset:
UNSET unset_terms
{
  $$ = algebra.NewUnset($2)
}
;

unset_terms:
unset_term
{
  $$ = algebra.UnsetTerms{$1}
}
|
unset_terms COMMA unset_term
{
  $$ = append($1, $3)
}
;

unset_term:
path opt_update_for
{
  $$ = algebra.NewUnsetTerm($1, $2)
}
;


/*************************************************
 *
 * MERGE
 *
 *************************************************/

merge:
MERGE INTO bucket_ref USING bucket_term ON key merge_actions opt_limit opt_returning
{
  $$ = algebra.NewMergeFrom($3, $5, "", $7, $8.Update, $8.Delete, $8.Insert, $9, $10)
}
|
MERGE INTO bucket_ref USING LPAREN from_term RPAREN as_alias ON key merge_actions opt_limit opt_returning
{
  $$ = algebra.NewMergeFrom($3, $6, $8, $10, $11.Update, $11.Delete, $11.Insert, $12, $13)
}
|
MERGE INTO bucket_ref USING LPAREN fullselect RPAREN as_alias ON key merge_actions opt_limit opt_returning
{
  $$ = algebra.NewMergeSelect($3, $6, $8, $10, $11.Update, $11.Delete, $11.Insert, $12, $13)
}
|
MERGE INTO bucket_ref USING LPAREN values RPAREN as_alias ON key merge_actions opt_limit opt_returning
{
  $$ = algebra.NewMergeValues($3, $6, $8, $10, $11.Update, $11.Delete, $11.Insert, $12, $13)
}
;

merge_actions:
/* empty */
{
  $$ = algebra.NewMergeActions(nil, nil, nil)
}
|
WHEN MATCHED THEN UPDATE merge_update opt_merge_delete_insert
{
  $$ = algebra.NewMergeActions($5, $6.Delete, $6.Insert)
}
|
WHEN MATCHED THEN DELETE merge_delete opt_merge_insert
{
  $$ = algebra.NewMergeActions(nil, $5, $6)
}
|
WHEN NOT MATCHED THEN INSERT merge_insert
{
  $$ = algebra.NewMergeActions(nil, nil, $6)
}
;

opt_merge_delete_insert:
/* empty */
{
  $$ = algebra.NewMergeActions(nil, nil, nil)
}
|
WHEN MATCHED THEN DELETE merge_delete opt_merge_insert
{
  $$ = algebra.NewMergeActions(nil, $5, $6)
}
|
WHEN NOT MATCHED THEN INSERT merge_insert
{
  $$ = algebra.NewMergeActions(nil, nil, $6)
}
;

opt_merge_insert:
/* empty */
{
  $$ = nil
}
|
WHEN NOT MATCHED THEN INSERT merge_insert
{
  $$ = $6
}
;

merge_update:
set opt_where
{
  $$ = algebra.NewMergeUpdate($1, nil, $2)
}
|
set unset opt_where
{
  $$ = algebra.NewMergeUpdate($1, $2, $3)
}
|
unset opt_where
{
  $$ = algebra.NewMergeUpdate(nil, $1, $2)
}
;

merge_delete:
opt_where
{
  $$ = algebra.NewMergeDelete($1)
}
;

merge_insert:
expr opt_where
{
  $$ = algebra.NewMergeInsert($1, $2)
}
;


/*************************************************
 *
 * CREATE INDEX
 *
 *************************************************/

create_index:
CREATE INDEX index_name ON named_bucket_ref LPAREN exprs RPAREN index_partition index_using
{
  $$ = algebra.NewCreateIndex($3, $5, $7, $9, $10)
}
;

index_name:
IDENTIFIER
;

named_bucket_ref:
bucket_name
{
  $$ = algebra.NewBucketRef("", $1, "")
}
|
pool_name COLON bucket_name
{
  $$ = algebra.NewBucketRef($1, $3, "")
}
;

index_partition:
/* empty */
{
  $$ = nil
}
|
PARTITION BY expr
{
  $$ = $3
}
;

index_using:
/* empty */
{
  $$ = catalog.VIEW
}
|
USING VIEW
{
  $$ = catalog.VIEW
}
;


/*************************************************
 *
 * DROP INDEX
 *
 *************************************************/

drop_index:
DROP INDEX named_bucket_ref DOT index_name
{
  $$ = algebra.NewDropIndex($3, $5)
}
;

/*************************************************
 *
 * ALTER INDEX
 *
 *************************************************/

alter_index:
ALTER INDEX named_bucket_ref DOT index_name rename
{
  $$ = algebra.NewAlterIndex($3, $5, $6)
}

rename:
/* empty */
{
  $$ = ""
}
|
RENAME TO index_name
{
  $$ = $3
}
;


/*************************************************
 *
 * Path
 *
 *************************************************/

path:
IDENTIFIER
{
  $$ = expression.NewIdentifier($1)
}
|
path DOT IDENTIFIER
{
  $$ = expression.NewField($1, expression.NewConstant(value.NewValue($3)))
}
|
path DOT LPAREN expr RPAREN
{
  $$ = expression.NewField($1, $4)
}
|
path LBRACKET expr RBRACKET
{
  $$ = expression.NewElement($1, $3)
}
;


/*************************************************
 *
 * Expression
 *
 *************************************************/

expr:
c_expr
|
/* Nested */
expr DOT IDENTIFIER
{
  $$ = expression.NewField($1, expression.NewConstant(value.NewValue($3)))
}
|
expr DOT LPAREN expr RPAREN
{
  $$ = expression.NewField($1, $4)
}
|
expr LBRACKET expr RBRACKET
{
  $$ = expression.NewElement($1, $3)
}
|
expr LBRACKET expr COLON RBRACKET
{
  $$ = expression.NewSlice($1, $3, nil)
}
|
expr LBRACKET expr COLON expr RBRACKET
{
  $$ = expression.NewSlice($1, $3, $5)
}
|
/* Arithmetic */
expr PLUS expr
{
  $$ = expression.NewAdd($1, $3)
}
|
expr MINUS expr
{
  $$ = expression.NewSubtract($1, $3)
}
|
expr STAR expr
{
  $$ = expression.NewMultiply($1, $3)
}
|
expr DIV expr
{
  $$ = expression.NewDivide($1, $3)
}
|
expr MOD expr
{
  $$ = expression.NewModulo($1, $3)
}
|
/* Concat */
expr CONCAT expr
{
  $$ = expression.NewConcat($1, $3)
}
|
/* Logical */
expr AND expr
{
  $$ = expression.NewAnd($1, $3)
}
|
expr OR expr
{
  $$ = expression.NewOr($1, $3)
}
|
NOT expr
{
  $$ = expression.NewNot($2)
}
|
/* Comparison */
expr EQ expr
{
  $$ = expression.NewEQ($1, $3)
}
|
expr DEQ expr
{
  $$ = expression.NewEQ($1, $3)
}
|
expr NE expr
{
  $$ = expression.NewNE($1, $3)
}
|
expr LT expr
{
  $$ = expression.NewLT($1, $3)
}
|
expr GT expr
{
  $$ = expression.NewGT($1, $3)
}
|
expr LE expr
{
  $$ = expression.NewLE($1, $3)
}
|
expr GE expr
{
  $$ = expression.NewGE($1, $3)
}
|
expr BETWEEN b_expr AND b_expr
{
  $$ = expression.NewBetween($1, $3, $5)
}
|
expr NOT BETWEEN b_expr AND b_expr
{
  $$ = expression.NewNotBetween($1, $4, $6)
}
|
expr LIKE expr
{
  $$ = expression.NewLike($1, $3)
}
|
expr NOT LIKE expr
{
  $$ = expression.NewNotLike($1, $4)
}
|
expr IN expr
{
  $$ = expression.NewIn($1, $3)
}
|
expr NOT IN expr
{
  $$ = expression.NewNotIn($1, $4)
}
|
expr IS NULL
{
  $$ = expression.NewIsNull($1)
}
|
expr IS NOT NULL
{
  $$ = expression.NewIsNotNull($1)
}
|
expr IS MISSING
{
  $$ = expression.NewIsMissing($1)
}
|
expr IS NOT MISSING
{
  $$ = expression.NewIsNotMissing($1)
}
|
expr IS VALUED
{
  $$ = expression.NewIsValued($1)
}
|
expr IS NOT VALUED
{
  $$ = expression.NewIsNotValued($1)
}
|
EXISTS expr
{
  $$ = expression.NewExists($2)
}
;

c_expr:
/* Literal */
literal
|
/* Identifier */
IDENTIFIER
{
  $$ = expression.NewIdentifier($1)
}
|
/* Function */
function_expr
|
/* Prefix */
MINUS expr %prec UMINUS
{
  $$ = expression.NewNegate($2)
}
|
/* Case */
case_expr
|
/* Collection */
collection_expr
|
/* Grouping and subquery */
paren_or_subquery_expr
;

b_expr:
c_expr
|
/* Nested */
b_expr DOT IDENTIFIER
{
  $$ = expression.NewField($1, expression.NewConstant(value.NewValue($3)))
}
|
b_expr DOT LPAREN expr RPAREN
{
  $$ = expression.NewField($1, $4)
}
|
b_expr LBRACKET expr RBRACKET
{
  $$ = expression.NewElement($1, $3)
}
|
b_expr LBRACKET expr COLON RBRACKET
{
  $$ = expression.NewSlice($1, $3, nil)
}
|
b_expr LBRACKET expr COLON expr RBRACKET
{
  $$ = expression.NewSlice($1, $3, $5)
}
|
/* Arithmetic */
b_expr PLUS b_expr
{
  $$ = expression.NewAdd($1, $3)
}
|
b_expr MINUS b_expr
{
  $$ = expression.NewSubtract($1, $3)
}
|
b_expr STAR b_expr
{
  $$ = expression.NewMultiply($1, $3)
}
|
b_expr DIV b_expr
{
  $$ = expression.NewDivide($1, $3)
}
|
b_expr MOD b_expr
{
  $$ = expression.NewModulo($1, $3)
}
|
/* Concat */
b_expr CONCAT b_expr
{
  $$ = expression.NewConcat($1, $3)
}
;


/*************************************************
 *
 * Literal
 *
 *************************************************/

literal:
NULL
{
  $$ = expression.NULL_EXPR
}
|
FALSE
{
  $$ = expression.FALSE_EXPR
}
|
TRUE
{
  $$ = expression.TRUE_EXPR
}
|
NUMBER
{
  $$ = expression.NewConstant(value.NewValue($1))
}
|
INT
{
  $$ = expression.NewConstant(value.NewValue($1))
}
|
STRING
{
  $$ = expression.NewConstant(value.NewValue($1))
}
|
object
|
array
;

object:
LBRACE opt_members RBRACE
{
  $$ = expression.NewObjectLiteral($2)
}
;

opt_members:
/* empty */
{
  $$ = nil
}
|
members
;

members:
member
{
  $$ = expression.Bindings{$1}
}
|
members COMMA member
{
  $$ = append($1, $3)
}
;

member:
STRING COLON expr
{
  $$ = expression.NewBinding($1, $3)
}
;

array:
LBRACKET opt_exprs RBRACKET
{
  $$ = expression.NewArrayLiteral($2)
}
;

opt_exprs:
/* empty */
{
  $$ = nil
}
|
exprs
;


/*************************************************
 *
 * Case
 *
 *************************************************/

case_expr:
CASE simple_or_searched_case END
{
  $$ = $2
}
;

simple_or_searched_case:
simple_case
|
searched_case
;

simple_case:
expr when_thens opt_else
{
  $$ = expression.NewSimpleCase($1, $2, $3)
}
;

when_thens:
WHEN expr THEN expr
{
  $$ = expression.WhenTerms{&expression.WhenTerm{$2, $4}}
}
|
when_thens WHEN expr THEN expr
{
  $$ = append($1, &expression.WhenTerm{$3, $5})
}
;

searched_case:
when_thens
opt_else
{
  $$ = expression.NewSearchedCase($1, $2)
}
;

opt_else:
/* empty */
{
  $$ = nil
}
|
ELSE expr
{
  $$ = $2
}
;


/*************************************************
 *
 * Function
 *
 *************************************************/

function_expr:
function_name LPAREN opt_exprs RPAREN
{
  $$ = nil;
  agg, ok := algebra.GetAggregate($1, false);
  if ok {
    $$ = agg.Constructor()($3);
  } else {
    f, ok := expression.GetFunction($1);
    if ok {
      $$ = f.Constructor()($3)
    }
  }
}
|
function_name LPAREN DISTINCT exprs RPAREN
{
  $$ = nil;
  agg, ok := algebra.GetAggregate($1, true);
  if ok {
      $$ = agg.Constructor()($4)
  }
}
|
function_name LPAREN STAR RPAREN
{
  $$ = nil;
  agg, ok := algebra.GetAggregate($1, false);
  if ok {
      $$ = agg.Constructor()(nil)
  }
}
;

function_name:
IDENTIFIER
;


/*************************************************
 *
 * Collection
 *
 *************************************************/

collection_expr:
collection_cond
|
collection_xform
;

collection_cond:
ANY coll_bindings satisfies END
{
  $$ = expression.NewAny($2, $3)
}
|
SOME coll_bindings satisfies END
{
  $$ = expression.NewAny($2, $3)
}
|
EVERY coll_bindings satisfies END
{
  $$ = expression.NewEvery($2, $3)
}
;

coll_bindings:
coll_binding
{
  $$ = expression.Bindings{$1}
}
|
coll_bindings COMMA coll_binding
{
  $$ = append($1, $3)
}
;

coll_binding:
variable IN expr
{
  $$ = expression.NewBinding($1, $3)
}
;

satisfies:
SATISFIES expr
{
  $$ = $2
}
;

collection_xform:
ARRAY expr FOR coll_bindings opt_when END
{
  $$ = expression.NewArray($2, $4, $5)
}
|
FIRST expr FOR coll_bindings opt_when END
{
  $$ = expression.NewFirst($2, $4, $5)
}
;


/*************************************************
 *
 * Parentheses and subquery
 *
 *************************************************/

paren_or_subquery_expr:
LPAREN paren_or_subquery RPAREN
{
  $$ = $2
}
;

paren_or_subquery:
expr
|
fullselect
{
  $$ = algebra.NewSubquery($1)
}
;