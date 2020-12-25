package util

import "github.com/sourcegraph/checkup/types"

func FindResultOfSameType(current *types.Result, list []types.Result) *types.Result {
	for _, result := range list {
		if result.Endpoint == current.Endpoint {
			return &result
		}
	}
	return nil
}
