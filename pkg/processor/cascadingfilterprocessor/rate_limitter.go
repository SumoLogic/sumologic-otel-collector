// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cascadingfilterprocessor

import "github.com/SumoLogic/sumologic-otel-collector/pkg/processor/cascadingfilterprocessor/sampling"

type rateLimiter struct {
	currentSecond        int64
	maxSpansPerSecond    int32
	spansInCurrentSecond int32
}

func newRateLimitter(maxSpansPerSecond int32) *rateLimiter {
	return &rateLimiter{
		maxSpansPerSecond: maxSpansPerSecond,
	}
}

// updateRate checks if given limit can still fit in the current limit
// returns Sampled when it's the case and NoTSampled otherwise
func (rl *rateLimiter) updateRate(currSecond int64, numSpans int32) sampling.Decision {
	// No limit equals no bounds
	if rl.maxSpansPerSecond <= 0 {
		return sampling.Sampled
	}

	if rl.currentSecond < currSecond {
		rl.currentSecond = currSecond
		rl.spansInCurrentSecond = 0
	}

	spansInSecondIfSampled := rl.spansInCurrentSecond + numSpans
	if spansInSecondIfSampled <= rl.maxSpansPerSecond {
		rl.spansInCurrentSecond = spansInSecondIfSampled
		return sampling.Sampled
	}

	return sampling.NotSampled
}
