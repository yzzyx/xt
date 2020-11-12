package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
)

type exprVal struct {
	Int        *IntValue
	Float      *IntValue
	String     *StringValue
	Identifier *Identifier
}

func compareEqual(lhs exprVal, rhs exprVal) bool {
	if lhs.Int != nil {
		if rhs.Int != nil {
			return lhs.Int.Val == rhs.Int.Val
		}
		if rhs.String != nil {
			rv, _ := strconv.Atoi(rhs.String.Val)
			return lhs.Int.Val == rv
		}
	}
	return false
}

func handleComparison(comparsion string, lhs exprVal, rhs exprVal) bool {
	switch comparsion {
	case "==":
		return compareEqual(lhs, rhs)
	}
	return false
}

// eval1Part handles expressions with only one node, effectively checking if a value is true
func eval1Part(expression []Node) (bool, error) {
	node := expression[0]
	switch e := node.(type) {
	case *IntValue:
		return e.Val != 0, nil
	case *StringValue:
		return strconv.ParseBool(e.Val)
		// FIXME - handle identifier
		//case *Identifier:
		//	ev.Identifier = e
	}
	return false, fmt.Errorf("unexpected node %s", node)
}

// eval3Parts handles comparisons
func eval3Parts(expression []Node) (bool, error) {
	comparison, ok := expression[1].(*Comparison)
	if !ok {
		return false, fmt.Errorf("expected comparison, got %s", expression[1])
	}

	var lhs, rhs exprVal
	switch e := expression[0].(type) {
	case *IntValue:
		lhs.Int = e
	case *StringValue:
		lhs.String = e
	case *Identifier:
		lhs.Identifier = e
	default:
		return false, fmt.Errorf("unexpected node %s", expression[0])
	}

	switch e := expression[2].(type) {
	case *IntValue:
		rhs.Int = e
	case *StringValue:
		rhs.String = e
	case *Identifier:
		rhs.Identifier = e
	default:
		return false, fmt.Errorf("unexpected node %s", expression[2])
	}

	ret := handleComparison(comparison.Type, lhs, rhs)
	return ret, nil
}

// EvaluateExpression  walks through 'expression' and returns true or false based on the expression
func EvaluateExpression(ctx context.Context, expression []Node) (bool, error) {

	// Expressions can be defined as:
	//  [not] <number/string/identifier/subexpression> [comparison <number/string/identifier/subexpression>]
	//     [or/and <expression>]...
	//
	// Which means that we can have the following parts, possibly separated by or/and:
	//  1 part:
	//     <number/string/identifier/subexpression>
	//  2 parts:
	//     not <number/string/identifier/subexpression>
	//  3 parts:
	//     <number/string/identifier/subexpression> <comparison> <number/string/identifier/subexpression>

	type Part struct {
		nodes []Node
	}

	// FIXME - split into and/or-parts
	parts := []Part{
		{nodes: expression},
	}

	var ret bool
	var err error
	for idx := range parts {

		switch len(parts[idx].nodes) {
		case 0:
			return false, fmt.Errorf("emtpy expression")
		case 1:
			ret, err = eval1Part(parts[idx].nodes)
		case 2:
			// FIXME - implement not-part
			//ret = eval2Part(parts[idx].nodes)
			return false, errors.New("support for 'not' is not implemented in expressions")
		case 3:
			ret, err = eval3Parts(parts[idx].nodes)
		}
		if err != nil {
			return false, err
		}
		return ret, nil
	}

	//idx := 0
	//next := func() Node {
	//	if idx == len(expression) {
	//		return nil
	//	}
	//	n := expression[idx]
	//	idx++
	//	return n
	//}
	return false, nil
}
