// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package invocationlifecycle

import (
	"os"
	"testing"
	"time"

	"github.com/DataDog/datadog-agent/pkg/trace/api"
	"github.com/DataDog/datadog-agent/pkg/trace/pb"
	"github.com/DataDog/datadog-agent/pkg/trace/sampler"
	"github.com/stretchr/testify/assert"
)

func TestStartExecutionSpanWithoutPayload(t *testing.T) {
	defer reset()
	startTime := time.Now()
	startExecutionSpan(startTime, "", LambdaInvokeEventHeaders{}, false)
	assert.Equal(t, startTime, currentExecutionInfo.startTime)
	assert.NotEqual(t, 0, currentExecutionInfo.traceID)
	assert.NotEqual(t, 0, currentExecutionInfo.spanID)
}

func TestStartExecutionSpanWithPayload(t *testing.T) {
	defer reset()
	testString := `a5a{"resource":"/users/create","path":"/users/create","httpMethod":"GET","headers":{"Accept":"*/*","Accept-Encoding":"gzip","x-datadog-parent-id":"1480558859903409531","x-datadog-sampling-priority":"1","x-datadog-trace-id":"5736943178450432258"}}0`
	startTime := time.Now()
	startExecutionSpan(startTime, testString, LambdaInvokeEventHeaders{}, false)
	assert.Equal(t, startTime, currentExecutionInfo.startTime)
	assert.Equal(t, uint64(5736943178450432258), currentExecutionInfo.traceID)
	assert.Equal(t, uint64(1480558859903409531), currentExecutionInfo.parentID)
	assert.Equal(t, sampler.SamplingPriority(1), *currentExecutionInfo.samplingPriority)
	assert.NotEqual(t, 0, currentExecutionInfo.spanID)
}

func TestStartExecutionSpanWithPayloadAndLambdaContextHeaders(t *testing.T) {
	defer reset()
	testString := `a5a{"resource":"/users/create","path":"/users/create","httpMethod":"GET"}0`
	lambdaInvokeContext := LambdaInvokeEventHeaders{
		TraceID:  "5736943178450432258",
		ParentID: "1480558859903409531",
	}
	startTime := time.Now()
	startExecutionSpan(startTime, testString, lambdaInvokeContext, false)
	assert.Equal(t, startTime, currentExecutionInfo.startTime)
	assert.Equal(t, uint64(5736943178450432258), currentExecutionInfo.traceID)
	assert.Equal(t, uint64(1480558859903409531), currentExecutionInfo.parentID)
	assert.NotEqual(t, 0, currentExecutionInfo.spanID)
}

func TestStartExecutionSpanWithPayloadAndInvalidIDs(t *testing.T) {
	defer reset()
	invalidTestString := `a5a{"resource":"/users/create","path":"/users/create","httpMethod":"GET","headers":{"Accept":"*/*","Accept-Encoding":"gzip","x-datadog-parent-id":"INVALID","x-datadog-sampling-priority":"1","x-datadog-trace-id":"INVALID"}}0`
	startTime := time.Now()
	startExecutionSpan(startTime, invalidTestString, LambdaInvokeEventHeaders{}, false)
	assert.Equal(t, startTime, currentExecutionInfo.startTime)
	assert.NotEqual(t, 9, currentExecutionInfo.traceID)
	assert.Equal(t, uint64(0), currentExecutionInfo.parentID)
	assert.NotEqual(t, 0, currentExecutionInfo.spanID)
}

func TestStartExecutionSpanWithNoHeadersAndInferredSpan(t *testing.T) {
	testString := `a5a{"resource":"/users/create","path":"/users/create","httpMethod":"GET"}0`
	startTime := time.Now()
	inferredSpan.Span = &pb.Span{
		TraceID: 2350923428932752492,
		SpanID:  1304592378509342580,
	}
	startExecutionSpan(startTime, testString, LambdaInvokeEventHeaders{}, true)
	assert.Equal(t, startTime, currentExecutionInfo.startTime)
	assert.Equal(t, uint64(2350923428932752492), currentExecutionInfo.traceID)
	assert.Equal(t, uint64(1304592378509342580), currentExecutionInfo.parentID)
	assert.NotEqual(t, 0, currentExecutionInfo.spanID)
}

func TestStartExecutionSpanWithHeadersAndInferredSpan(t *testing.T) {
	testString := `a5a{"resource":"/users/create","path":"/users/create","httpMethod":"GET","headers":{"Accept":"*/*","Accept-Encoding":"gzip","x-datadog-parent-id":"1480558859903409531","x-datadog-sampling-priority":"1","x-datadog-trace-id":"5736943178450432258"}}0`
	startTime := time.Now()
	inferredSpan.Span = &pb.Span{
		SpanID: 1304592378509342580,
	}
	startExecutionSpan(startTime, testString, LambdaInvokeEventHeaders{}, true)
	assert.Equal(t, startTime, currentExecutionInfo.startTime)
	assert.Equal(t, uint64(5736943178450432258), currentExecutionInfo.traceID)
	assert.Equal(t, uint64(1304592378509342580), currentExecutionInfo.parentID)
	assert.Equal(t, sampler.SamplingPriority(1), *currentExecutionInfo.samplingPriority)
	assert.Equal(t, uint64(5736943178450432258), inferredSpan.Span.TraceID)
	assert.Equal(t, uint64(1480558859903409531), inferredSpan.Span.ParentID)
	assert.Equal(t, sampler.SamplingPriority(1), *inferredSpan.SamplingPriority)

	assert.NotEqual(t, 0, currentExecutionInfo.spanID)
}
func TestEndExecutionSpanWithNoError(t *testing.T) {
	defer os.Unsetenv(functionNameEnvVar)
	defer os.Unsetenv("DD_CAPTURE_LAMBDA_PAYLOAD")
	os.Setenv(functionNameEnvVar, "TestFunction")
	os.Setenv("DD_CAPTURE_LAMBDA_PAYLOAD", "true")
	defer reset()
	testString := `a5a{"resource":"/users/create","path":"/users/create","httpMethod":"GET","headers":{"Accept":"*/*","Accept-Encoding":"gzip","x-datadog-parent-id":"1480558859903409531","x-datadog-sampling-priority":"1","x-datadog-trace-id":"5736943178450432258"}}0`
	startTime := time.Now()
	startExecutionSpan(startTime, testString, LambdaInvokeEventHeaders{}, false)

	duration := 1 * time.Second
	endTime := startTime.Add(duration)
	isError := false
	var tracePayload *api.Payload
	mockProcessTrace := func(payload *api.Payload) {
		tracePayload = payload
	}

	endExecutionSpan(mockProcessTrace, "test-request-id", endTime, isError, []byte(`{"response":"test response payload"}`))
	executionSpan := tracePayload.TracerPayload.Chunks[0].Spans[0]
	assert.Equal(t, "aws.lambda", executionSpan.Name)
	assert.Equal(t, "aws.lambda", executionSpan.Service)
	assert.Equal(t, "TestFunction", executionSpan.Resource)
	assert.Equal(t, "serverless", executionSpan.Type)
	assert.Equal(t, "test-request-id", executionSpan.Meta["request_id"])
	assert.Equal(t, testString, executionSpan.Meta["function.request"])
	assert.Equal(t, `{"response":"test response payload"}`, executionSpan.Meta["function.response"])
	assert.Equal(t, currentExecutionInfo.traceID, executionSpan.TraceID)
	assert.Equal(t, currentExecutionInfo.spanID, executionSpan.SpanID)
	assert.Equal(t, startTime.UnixNano(), executionSpan.Start)
	assert.Equal(t, duration.Nanoseconds(), executionSpan.Duration)
}

func TestEndExecutionSpanWithInvalidCaptureLambdaPayloadValue(t *testing.T) {
	defer os.Unsetenv(functionNameEnvVar)
	defer os.Unsetenv("DD_CAPTURE_LAMBDA_PAYLOAD")
	os.Setenv(functionNameEnvVar, "TestFunction")
	os.Setenv("DD_CAPTURE_LAMBDA_PAYLOAD", "INVALID_INPUT")
	defer reset()
	testString := `a5a{"resource":"/users/create","path":"/users/create","httpMethod":"GET","headers":{"Accept":"*/*","Accept-Encoding":"gzip","x-datadog-parent-id":"1480558859903409531","x-datadog-sampling-priority":"1","x-datadog-trace-id":"5736943178450432258"}}0`
	startTime := time.Now()
	startExecutionSpan(startTime, testString, LambdaInvokeEventHeaders{}, false)

	duration := 1 * time.Second
	endTime := startTime.Add(duration)
	isError := false
	var tracePayload *api.Payload
	mockProcessTrace := func(payload *api.Payload) {
		tracePayload = payload
	}

	endExecutionSpan(mockProcessTrace, "test-request-id", endTime, isError, []byte(`{"response":"test response payload"}`))
	executionSpan := tracePayload.TracerPayload.Chunks[0].Spans[0]
	assert.Equal(t, "aws.lambda", executionSpan.Name)
	assert.Equal(t, "aws.lambda", executionSpan.Service)
	assert.Equal(t, "TestFunction", executionSpan.Resource)
	assert.Equal(t, "serverless", executionSpan.Type)
	assert.Equal(t, "test-request-id", executionSpan.Meta["request_id"])
	assert.NotContains(t, executionSpan.Meta, "function.request")
	assert.NotContains(t, executionSpan.Meta, "function.response")
	assert.Equal(t, currentExecutionInfo.traceID, executionSpan.TraceID)
	assert.Equal(t, currentExecutionInfo.spanID, executionSpan.SpanID)
	assert.Equal(t, startTime.UnixNano(), executionSpan.Start)
	assert.Equal(t, duration.Nanoseconds(), executionSpan.Duration)
}

func TestEndExecutionSpanWithError(t *testing.T) {
	defer os.Unsetenv(functionNameEnvVar)
	os.Setenv(functionNameEnvVar, "TestFunction")
	defer reset()
	testString := `a5a{"resource":"/users/create","path":"/users/create","httpMethod":"GET","headers":{"Accept":"*/*","Accept-Encoding":"gzip","x-datadog-parent-id":"1480558859903409531","x-datadog-sampling-priority":"1","x-datadog-trace-id":"5736943178450432258"}}0`
	startTime := time.Now()
	startExecutionSpan(startTime, testString, LambdaInvokeEventHeaders{}, false)

	duration := 1 * time.Second
	endTime := startTime.Add(duration)
	isError := true
	var tracePayload *api.Payload
	mockProcessTrace := func(payload *api.Payload) {
		tracePayload = payload
	}

	endExecutionSpan(mockProcessTrace, "test-request-id", endTime, isError, []byte("{}"))
	executionSpan := tracePayload.TracerPayload.Chunks[0].Spans[0]
	assert.Equal(t, executionSpan.Error, int32(1))
}

func TestConvertRawPayloadWithHeaders(t *testing.T) {

	s := `a5a{"resource":"/users/create","path":"/users/create","httpMethod":"GET","headers":{"Accept":"*/*","Accept-Encoding":"gzip","x-datadog-parent-id":"1480558859903409531","x-datadog-sampling-priority":"1","x-datadog-trace-id":"5736943178450432258"}}0`

	expectedPayload := invocationPayload{}
	expectedPayload.Headers = map[string]string{"Accept": "*/*", "Accept-Encoding": "gzip", "x-datadog-parent-id": "1480558859903409531", "x-datadog-sampling-priority": "1", "x-datadog-trace-id": "5736943178450432258"}

	p := convertRawPayload(s)

	assert.Equal(t, p, expectedPayload)
}

func TestConvertRawPayloadWithOutHeaders(t *testing.T) {

	s := `a5a{"resource":"/users/create","path":"/users/create","httpMethod":"GET"}0`

	expectedPayload := invocationPayload{}

	p := convertRawPayload(s)

	assert.Equal(t, p, expectedPayload)
}

func reset() {
	currentExecutionInfo = executionStartInfo{}
}
