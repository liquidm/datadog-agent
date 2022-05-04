// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package metrics

import (
	"strings"
	"testing"

	"github.com/DataDog/datadog-agent/pkg/metrics"
	"github.com/stretchr/testify/require"
)

func TestIterableSeriesEmptyMarshalJSON(t *testing.T) {
	r := require.New(t)
	iterableSerie := metrics.NewIterableSeries(func(*metrics.Serie) {}, 10, 2)
	iterableSeriesSerializer := IterableSeries{SerieSource: iterableSerie}
	iterableSerie.SenderStopped()
	bytes, err := iterableSeriesSerializer.MarshalJSON()
	r.NoError(err)
	r.Equal(`{"series":[]}`, strings.TrimSpace(string(bytes)))
}
