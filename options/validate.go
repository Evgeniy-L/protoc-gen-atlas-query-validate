package options

import (
	"fmt"
	"strings"

	"github.com/infobloxopen/atlas-app-toolkit/query"
)

type FilteringOptions struct {
	AllowMissingFields bool
	Options            map[string]FilteringOption
}

type FilteringOption struct {
	DisableSorting bool
	FilterType     QueryValidate_FilterType
	Deny           []QueryValidate_FilterOperator
}

func ValidateFiltering(f *query.Filtering, messageInfo map[string]FilteringOption) error {
	var getOperator func(interface{}) error
	var fieldInfo FilteringOption

	validate := func(path []string, f interface{}) error {

		var ok bool
		fieldTag := strings.Join(path, ".")
		fieldInfo, ok = messageInfo[fieldTag]
		if !ok {
			return fmt.Errorf("Unknown field: %s", fieldTag)
		}

		tp := ""

		switch x := f.(type) {
		case *query.StringCondition:
			if fieldInfo.FilterType != QueryValidate_STRING {
				return fmt.Errorf("Got invalid literal type for %s, expect %s", fieldTag, fieldInfo.FilterType)
			}
			sc := &query.Filtering_StringCondition{x}
			tp = query.StringCondition_Type_name[int32(sc.StringCondition.Type)]
		case *query.NumberCondition:
			if fieldInfo.FilterType != QueryValidate_NUMBER {
				return fmt.Errorf("Got invalid literal type for %s, expect %s", fieldTag, fieldInfo.FilterType)
			}
			nc := &query.Filtering_NumberCondition{x}
			tp = query.NumberCondition_Type_name[int32(nc.NumberCondition.Type)]
		default:
			return nil
		}
		for _, val := range fieldInfo.Deny {
			if val == QueryValidate_ALL {
				return fmt.Errorf("Operation %s is not allowed for '%s'", tp, fieldTag)
			}
			if val.String() == tp {
				return fmt.Errorf("Operation %s is not allowed for '%s'", tp, fieldTag)
			}
		}
		return nil
	}

	var vres error

	getOperator = func(f interface{}) error {
		val := f.(*query.LogicalOperator)
		var vres error
		left := val.GetLeft()
		switch leftVal := left.(type) {
		case *query.LogicalOperator_LeftOperator:
			vres = getOperator(leftVal.LeftOperator)

		case *query.LogicalOperator_LeftStringCondition:
			vres = validate(leftVal.LeftStringCondition.GetFieldPath(), leftVal.LeftStringCondition)

		case *query.LogicalOperator_LeftNumberCondition:
			vres = validate(leftVal.LeftNumberCondition.GetFieldPath(), leftVal.LeftNumberCondition)

		case *query.LogicalOperator_LeftNullCondition:
			vres = validate(leftVal.LeftNullCondition.GetFieldPath(), leftVal.LeftNullCondition)
		}

		if vres != nil {
			return vres
		}

		right := val.GetRight()
		switch rightVal := right.(type) {
		case *query.LogicalOperator_RightOperator:
			getOperator(rightVal.RightOperator)

		case *query.LogicalOperator_RightStringCondition:
			vres = validate(rightVal.RightStringCondition.GetFieldPath(), rightVal.RightStringCondition)

		case *query.LogicalOperator_RightNumberCondition:
			vres = validate(rightVal.RightNumberCondition.GetFieldPath(), rightVal.RightNumberCondition)

		case *query.LogicalOperator_RightNullCondition:
			vres = validate(rightVal.RightNullCondition.GetFieldPath(), rightVal.RightNullCondition)
		}

		return vres
	}

	if f != nil {
		root := f.GetRoot()
		switch val := root.(type) {
		case *query.Filtering_Operator:
			vres = getOperator(val.Operator)

		case *query.Filtering_StringCondition:
			vres = validate(val.StringCondition.GetFieldPath(), val.StringCondition)

		case *query.Filtering_NumberCondition:
			vres = validate(val.NumberCondition.GetFieldPath(), val.NumberCondition)

		case *query.Filtering_NullCondition:
			vres = validate(val.NullCondition.GetFieldPath(), val.NullCondition)
		}
	}
	return vres
}

func ValidateSorting(p *query.Sorting, messageInfo map[string]FilteringOption) error {
	var res error
	var fieldInfo FilteringOption

	if p != nil {
		for _, criteria := range p.GetCriterias() {
			tag := criteria.GetTag()
			var ok bool
			fieldInfo, ok = messageInfo[tag]
			if !ok {
				return fmt.Errorf("Unknown field: %s", tag)
			}

			if fieldInfo.DisableSorting {
				res = fmt.Errorf("Sorting is not allowed for '%s'", tag)
				break
			}
		}
	}
	return res
}