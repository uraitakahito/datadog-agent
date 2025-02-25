// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// Package rules holds rules related files
package rules

import (
	"errors"

	"github.com/DataDog/datadog-agent/pkg/security/secl/compiler/eval"
)

// Approvers are just filter values indexed by field
type Approvers map[eval.Field]FilterValues

// isAnApprover returns whether the given value is an approver for the given rule
func isAnApprover(event eval.Event, ctx *eval.Context, rule *Rule, field eval.Field, value interface{}) (bool, error) {
	var readOnlyError *eval.ErrFieldReadOnly
	if err := event.SetFieldValue(field, value); err != nil {
		if errors.As(err, &readOnlyError) {
			return false, nil
		}
		return false, err
	}
	origResult, err := rule.PartialEval(ctx, field)
	if err != nil {
		return false, err
	}

	notValue, err := eval.NotOfValue(value)
	if err != nil {
		return false, err
	}

	if err := event.SetFieldValue(field, notValue); err != nil {
		if errors.As(err, &readOnlyError) {
			return false, nil
		}
		return false, err
	}
	notResult, err := rule.PartialEval(ctx, field)
	if err != nil {
		return false, err
	}

	return origResult && !notResult, nil
}

func bitmaskCombinations(bitmasks []int) []int {
	if len(bitmasks) == 0 {
		return nil
	}

	combinationCount := 1 << len(bitmasks)
	result := make([]int, 0, combinationCount)
	for i := 0; i < combinationCount; i++ {
		var mask int
		for j, value := range bitmasks {
			if (i & (1 << j)) > 0 {
				mask |= value
			}
		}
		result = append(result, mask)
	}

	return result
}

//  1. all the rule for a given event type has to have approvers
//     with:
//     * caps: open.file.name # only able to apply approver for open.file.name, not for open.flags
//     ok:
//     * open.file.name == "123" && process.uid == 33
//     * open.file.name == "567" && process.gid == 55
//     ko:
//     * open.file.name == "123" && process.uid == 33
//     * open.flags == O_RDONLY
//     reason:
//     * We can let pass only the event for the `open.file.name` of the first rule as the second one has to be evaluated on all the open events.
//
//  2. all the approver values has to be captured and used by the in-kernel filtering mechanism
//     ex:
//     * open.file.name in ["123", "456"] && process.uid == 33
//     * open.file.name == "567" && process.gid == 55
//     => approver("123", "456", "567")
//
//  3. non approver values can co-exists with approver value in the same rule
//     ex:
//     * open.file.name in ["123", "456"] && open.file.name != "4.*" && open.file.name != "888"
//     reason:
//     * event will be approved kernel side and will be rejected userspace side
func getApprovers(rules []*Rule, event eval.Event, fieldCaps FieldCapabilities) (Approvers, error) {
	approvers := make(Approvers)

	ctx := eval.NewContext(event)

	// for each rule we should at least find one approver otherwise we will return no approver for the field
	for _, rule := range rules {
		var (
			bestFilterField  eval.Field
			bestFilterValues FilterValues
			bestFilterWeight int
			bestFilterMode   FilterMode
		)

		for _, fieldCap := range fieldCaps {
			field := fieldCap.Field

			var filterValues FilterValues
			var bitmasks []int

			for _, value := range rule.GetFieldValues(field) {
				switch value.Type {
				case eval.ScalarValueType, eval.PatternValueType, eval.GlobValueType:
					isAnApprover, err := isAnApprover(event, ctx, rule, field, value.Value)
					if err != nil {
						return nil, err
					}

					if isAnApprover {
						filterValues = filterValues.Merge(FilterValue{Field: field, Value: value.Value, Type: value.Type, Mode: fieldCap.FilterMode})
					}
				case eval.BitmaskValueType:
					bitmasks = append(bitmasks, value.Value.(int))
				}
			}

			for _, bitmask := range bitmaskCombinations(bitmasks) {
				isAnApprover, err := isAnApprover(event, ctx, rule, field, bitmask)
				if err != nil {
					return nil, err
				}

				if isAnApprover {
					filterValues = filterValues.Merge(FilterValue{Field: field, Value: bitmask, Type: eval.BitmaskValueType})
				}
			}

			if len(filterValues) == 0 || !fieldCaps.Validate(filterValues) {
				continue
			}

			if bestFilterValues == nil || fieldCap.FilterWeight > bestFilterWeight {
				bestFilterField = field
				bestFilterValues = filterValues
				bestFilterWeight = fieldCap.FilterWeight
				bestFilterMode = fieldCap.FilterMode
			}
		}

		// no filter value for a rule thus no approver for the event type
		if bestFilterValues == nil {
			return nil, nil
		}

		// this rule as an approver in ApproverOnly mode. Disable to rule from being used by the discarder mechanism.
		// the goal is to have approver that are good enough to filter properly the events used by rule that would break the
		// discarder discovery.
		// ex: open.file.name != "" && process.auid == 1000 # this rule avoid open.file.path discarder discovery, but with a ApproverOnly on process.auid the problem disappear
		//     open.file.filename == "/etc/passwd"
		if bestFilterMode == ApproverOnlyMode {
			rule.NoDiscarder = true
		}

		approvers[bestFilterField] = append(approvers[bestFilterField], bestFilterValues...)
	}

	return approvers, nil
}
