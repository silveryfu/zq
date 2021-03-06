package driver

import (
	"testing"

	"github.com/brimsec/zq/ast"
	"github.com/brimsec/zq/zql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeColumns(t *testing.T) {
	tests := []struct {
		zql  string
		cols []string
	}{
		{
			"cut x, y, z",
			[]string{"x", "y", "z"},
		},
		{
			"cut -c foo | cut x, y, z",
			[]string{"x", "y", "z"},
		},
		{
			"put x=y | filter x > 1",
			nil,
		},
		{
			"put x=y | filter x > 1 | cut x",
			[]string{"x", "y"},
		},
		{
			"head 1 | tail 1 | cut x, y, z",
			[]string{"x", "y", "z"},
		},
		{
			"count()",
			[]string{},
		},
		{
			"tail | count()",
			[]string{},
		},
		{
			"count(y) by x",
			[]string{"x", "y"},
		},
		{
			"every 1s count(y) by x",
			[]string{"ts", "x", "y"},
		},
		{
			"every 1s count(y) by foo=String.replace(x, y, z)",
			[]string{"ts", "x", "y", "z"},
		},
		{
			"filter x=1 | every 1s count(y) by foo=z | (head 1; tail 1)",
			[]string{"ts", "x", "y", "z"},
		},
		{
			"put x=1 | every 1s count(y) by foo=z | (head 1; tail 1)",
			[]string{"ts", "y", "z"},
		},
		{
			"rename foo=x | every 1s count(y) by foo=z | (head 1; tail 1)",
			[]string{"ts", "x", "y", "z"},
		},
		{
			"*>1 | every 1s count(y) by foo=String.replace(x, y, z) | (head 1; tail 1)",
			nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.zql, func(t *testing.T) {
			query, err := zql.ParseProc(tc.zql)
			require.NoError(t, err)
			// apply ReplaceGroupByProcDurationWithKey here because computeColumns
			// will be applied after it when it is plugged into compilation.
			ReplaceGroupByProcDurationWithKey(query)
			cols := computeColumns(query)
			var expected map[string]struct{}
			if tc.cols != nil {
				expected = make(map[string]struct{})
				for _, s := range tc.cols {
					expected[s] = struct{}{}
				}
			}
			assert.Equal(t, expected, cols)
		})
	}
}

func TestExpressionFields(t *testing.T) {
	tests := []struct {
		expr     string
		expected []string
	}{
		{
			"a=1",
			[]string{"a"},
		},
		{
			"a:time",
			[]string{"a"},
		},
		{
			"!a=1",
			[]string{"a"},
		},
		{
			"a=1 or b=c",
			[]string{"a", "b", "c"},
		},
		{
			"a ? b : c + d",
			[]string{"a", "b", "c", "d"},
		},
		{
			"Time.trunc(ts, 10)",
			[]string{"ts"},
		},
		{
			"Math.max(a, b, c, d)",
			[]string{"a", "b", "c", "d"},
		},
		{
			"String.replace(a, b, c)",
			[]string{"a", "b", "c"},
		},
		{
			"String.replace(a, 'b', c)",
			[]string{"a", "c"},
		},
		{
			"x = 0 ? y : z",
			[]string{"x", "y", "z"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.expr, func(t *testing.T) {
			e, err := zql.ParseExpression(tc.expr)
			require.NoError(t, err)
			f := expressionFields(e)
			assert.Equal(t, tc.expected, f)
		})
	}
}

func TestBooleanExpressionFields(t *testing.T) {
	tests := []struct {
		expr     string
		expected []string
	}{
		{
			"* = 1",
			nil,
		},
		{
			"*",
			[]string{},
		},
		{
			"* OR *=1",
			nil,
		},
		{
			"* OR x=1",
			[]string{"x"},
		},
		{
			"!(* OR x=1)",
			[]string{"x"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.expr, func(t *testing.T) {
			parsed, err := zql.ParseProc(tc.expr)
			require.NoError(t, err)
			f := booleanExpressionFields(parsed.(*ast.FilterProc).Filter)
			assert.Equal(t, tc.expected, f)
		})
	}
}

func TestParallelizeFlowgraph(t *testing.T) {
	tests := []struct {
		zql                     string
		orderField              string
		expected                string
		expectedMergeOrderField string
	}{
		{
			"* | uniq",
			"ts",
			"(filter *; filter *) | uniq",
			"ts",
		},
		{
			"* | cut ts, foo=x | uniq",
			"ts",
			"(filter * | cut ts, foo=x; filter * | cut ts, foo=x) | uniq",
			"ts",
		},
		{
			"* | cut -c x | uniq",
			"ts",
			"(filter * | cut -c x; filter * | cut -c x) | uniq",
			"ts",
		},
		{
			"* | put ts=foo | rename foo=boo",
			"ts",
			"(filter *; filter *) | put ts=foo | rename foo=boo",
			"ts",
		},
		{
			"* | put ts=foo | rename foo=boo | count()",
			"ts",
			"(filter * | put ts=foo | rename foo=boo | count(); filter * | put ts=foo | rename foo=boo | count()) | count()",
			"",
		},
		{
			"* | put ts=foo | rename foo=boo | sort",
			"ts",
			"(filter * | put ts=foo | rename foo=boo; filter * | put ts=foo | rename foo=boo) | sort",
			"",
		},
		{
			"* | put x=foo | rename foo=boo | uniq",
			"ts",
			"(filter * | put x=foo | rename foo=boo; filter * | put x=foo | rename foo=boo) | uniq",
			"ts",
		},
		{
			"* | sort x | uniq",
			"ts",
			"(filter * | sort x; filter * | sort x) | uniq",
			"x",
		},
		{
			"* | sort | uniq",
			"ts",
			"(filter *; filter *) | sort | uniq",
			"",
		},
		{
			"* | put x=y | countdistinct(x) by y | uniq",
			"ts",
			"(filter * | put x=y; filter * | put x=y) | countdistinct(x) by y | uniq",
			"ts",
		},
		{
			"* | count() by y",
			"ts",
			"(filter * | count() by y; filter * | count() by y) | count() by y",
			"",
		},
		{
			"* | every 1h count() by y",
			"",
			"(filter * | every 1h count() by y; filter * | every 1h count() by y) | every 1h count() by y",
			"ts",
		},
		{
			"* | put a=1 | tail",
			"ts",
			"(filter * | put a=1 | tail; filter * | put a=1 | tail) | tail",
			"ts",
		},
	}
	for _, tc := range tests {
		t.Run(tc.zql, func(t *testing.T) {
			query, err := zql.ParseProc(tc.zql)
			require.NoError(t, err)
			parallelized, ok := parallelizeFlowgraph(query.(*ast.SequentialProc), 2, tc.orderField, false)
			require.Equal(t, ok, tc.zql != tc.expected)

			expected, err := zql.ParseProc(tc.expected)
			require.NoError(t, err)
			if _, ok := expected.(*ast.SequentialProc).Procs[1].(*ast.ParallelProc); ok {
				// Remove the "filter *" that is pre-pended during parsing
				// if the proc started with a parallel graph.
				expected.(*ast.SequentialProc).Procs = expected.(*ast.SequentialProc).Procs[1:]

				// If the parallelized flowgraph includes a groupby, then adjust the expected AST by setting
				// the EmitPart flag on the parallelized groupbys and the ConsumePart flag on the post-merge groupby.
				branch := expected.(*ast.SequentialProc).Procs[0].(*ast.ParallelProc).Procs[0].(*ast.SequentialProc)
				if _, ok := branch.Procs[len(branch.Procs)-1].(*ast.GroupByProc); ok {
					for _, b := range expected.(*ast.SequentialProc).Procs[0].(*ast.ParallelProc).Procs {
						seq := b.(*ast.SequentialProc)
						seq.Procs[len(seq.Procs)-1].(*ast.GroupByProc).EmitPart = true
					}
					g := expected.(*ast.SequentialProc).Procs[1].(*ast.GroupByProc)
					g.ConsumePart = true
				}
			}
			if tc.zql != tc.expected {
				expected.(*ast.SequentialProc).Procs[0].(*ast.ParallelProc).MergeOrderField = tc.expectedMergeOrderField
			}
			assert.Equal(t, expected.(*ast.SequentialProc), parallelized)
		})
	}
	// This needs a standalone test due to the presence of a pass
	// proc in the transformed AST.
	t.Run("* | cut ts, y, z | put x=y | rename y=z", func(t *testing.T) {
		orderField := "ts"
		query := "* | cut ts, y, z | put x=y | rename y=z"
		dquery := "(filter * | cut ts, y, z | put x=y | rename y=z; filter * | cut ts, y, z | put x=y | rename y=z)"
		program, err := zql.ParseProc(query)
		require.NoError(t, err)
		parallelized, ok := parallelizeFlowgraph(program.(*ast.SequentialProc), 2, orderField, false)
		require.True(t, ok)

		expected, err := zql.ParseProc(dquery)
		require.NoError(t, err)

		// We can't express a pass proc in zql, so add it to the AST this way.
		// (It's added by the parallelized flowgraph in order to force a merge rather than having trailing leaves connected to a mux output).
		expected.(*ast.SequentialProc).Procs = append(expected.(*ast.SequentialProc).Procs[1:], &ast.PassProc{Node: ast.Node{"PassProc"}})
		expected.(*ast.SequentialProc).Procs[0].(*ast.ParallelProc).MergeOrderField = orderField

		assert.Equal(t, expected.(*ast.SequentialProc), parallelized)
	})
}

func TestSetGroupByProcInputSortDir(t *testing.T) {
	tests := []struct {
		zql            string
		inputSortField string
		groupbySortDir int
		outputSorted   bool
	}{
		{
			"* | every 1h count()",
			"ts",
			1,
			true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.zql, func(t *testing.T) {
			query, err := zql.ParseProc(tc.zql)
			require.NoError(t, err)
			ReplaceGroupByProcDurationWithKey(query)
			outputSorted := setGroupByProcInputSortDir(query, tc.inputSortField, 1)
			require.Equal(t, tc.outputSorted, outputSorted)

			var found bool
			for _, b := range query.(*ast.SequentialProc).Procs {
				if gbp, ok := b.(*ast.GroupByProc); ok {
					require.Equal(t, tc.groupbySortDir, gbp.InputSortDir)
					found = true
					break
				}
			}
			require.True(t, found)
		})
	}
}
