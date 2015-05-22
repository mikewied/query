//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
    "runtime/debug"

	"github.com/couchbase/query/errors"
	"github.com/couchbase/query/execution"
	"github.com/couchbase/query/server"
	"github.com/couchbase/query/value"
)

func (this *httpRequest) Output() execution.Output {
	return this
}

func (this *httpRequest) Fail(err errors.Error) {
	this.SetState(server.FATAL)
	// Determine the appropriate http response code based on the error
	httpRespCode := mapErrorToHttpResponse(err)
	this.setHttpCode(httpRespCode)
	// Put the error on the errors channel
	this.Errors() <- err
}

func mapErrorToHttpResponse(err errors.Error) int {
	switch err.Code() {
	case 1000: // readonly violation
		return http.StatusForbidden
	case 1010: // unsupported http method
		return http.StatusMethodNotAllowed
	case 1020, 1030, 1040, 1050, 1060, 1070:
		return http.StatusBadRequest
	case 1120:
		return http.StatusNotAcceptable
	case 3000: // parse error range
		return http.StatusBadRequest
	case 4000: // plan error range
		return http.StatusNotFound
	case 5000:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func (this *httpRequest) httpCode() int {
	this.RLock()
	defer this.RUnlock()
	return this.httpRespCode
}

func (this *httpRequest) setHttpCode(httpRespCode int) {
	this.Lock()
	defer this.Unlock()
	this.httpRespCode = httpRespCode
}

func (this *httpRequest) Failed(srvr *server.Server) {
	defer this.stopAndClose(server.FATAL)

	this.writeString("{\n")
	this.writeRequestID()
	this.writeClientContextID()
	this.writeErrors()
	this.writeWarnings()
	this.writeState("")
	this.writeMetrics(srvr.Metrics())
	this.writeString("\n}\n")
	this.writer.noMoreData()
}

func (this *httpRequest) Execute(srvr *server.Server, signature value.Value, stopNotify chan bool) {
	defer this.stopAndClose(server.COMPLETED)

	this.NotifyStop(stopNotify)

	this.setHttpCode(http.StatusOK)
	_ = this.writePrefix(srvr, signature) &&
		this.writeResults()
	this.writeSuffix(srvr.Metrics(), "")
	this.writer.noMoreData()
}

func (this *httpRequest) Expire() {
	defer this.stopAndClose(server.TIMEOUT)

	if this.httpCode() == 0 {
		this.setHttpCode(http.StatusOK)
		this.writePrefix(&server.Server{}, nil)
	}
	this.writeSuffix(true, server.TIMEOUT)
	this.writer.noMoreData()
}

func (this *httpRequest) stopAndClose(state server.State) {
	this.Stop(state)
	this.stopCloseNotifier()
	this.Close()
}

func (this *httpRequest) stopCloseNotifier() {
	select {
	case this.requestNotify <- false:
	default:
	}
}

func (this *httpRequest) writePrefix(srvr *server.Server, signature value.Value) bool {
	return this.writeString("{\n") &&
		this.writeRequestID() &&
		this.writeClientContextID() &&
		this.writeSignature(srvr.Signature(), signature) &&
		this.writeString(",\n    \"results\": [")
}

func (this *httpRequest) writeRequestID() bool {
	return this.writeString(fmt.Sprintf("    \"requestID\": \"%s\"", this.Id().String()))
}

func (this *httpRequest) writeClientContextID() bool {
	if !this.ClientID().IsValid() {
		return true
	}
	return this.writeString(fmt.Sprintf(",\n    \"clientContextID\": \"%s\"", this.ClientID().String()))
}

func (this *httpRequest) writeSignature(server_flag bool, signature value.Value) bool {
	s := this.Signature()
	if s == value.FALSE ||
		(s == value.NONE && !server_flag) {
		return true
	}
	return this.writeString(",\n    \"signature\": ") &&
		this.writeValue(signature)
}

func (this *httpRequest) writeResults() bool {
	var item value.Value

	ok := true
	for ok {
		select {
		case <-this.StopExecute():
			this.SetState(server.STOPPED)
			return true
		default:
		}

		select {
		case item, ok = <-this.Results():
			if ok {
				if !this.writeResult(item) {
					this.SetState(server.FATAL)
					return false
				}
			}
		case <-this.StopExecute():
			this.SetState(server.STOPPED)
			return true
		}
	}

	this.SetState(server.COMPLETED)
	return true
}

func (this *httpRequest) writeResult(item value.Value) bool {
	var rv bool
	if this.resultCount == 0 {
		rv = this.writeString("\n")
	} else {
		rv = this.writeString(",\n")
	}

	bytes, err := json.MarshalIndent(item, "        ", "    ")
	if err != nil {
		this.Errors() <- errors.NewServiceErrorInvalidJSON(err)
		return false
	}

	this.resultSize += len(bytes)
	this.resultCount++

	return rv &&
		this.writeString("        ") &&
		this.writeString(string(bytes))
}

func (this *httpRequest) writeValue(item value.Value) bool {
	bytes, err := json.MarshalIndent(item, "    ", "    ")
	if err != nil {
		panic(err.Error())
	}

	return this.writeString(string(bytes))
}

func (this *httpRequest) writeSuffix(metrics bool, state server.State) bool {
	return this.writeString("\n    ]") &&
		this.writeErrors() &&
		this.writeWarnings() &&
		this.writeState(state) &&
		this.writeMetrics(metrics) &&
		this.writeString("\n}\n")
}

func (this *httpRequest) writeString(s string) bool {
	return this.writer.writeString(s)
}

func (this *httpRequest) writeState(state server.State) bool {
	if state == "" {
		state = this.State()
	}

	if state == server.COMPLETED {
		if this.errorCount == 0 {
			state = server.SUCCESS
		} else {
			state = server.ERRORS
		}
	}

	return this.writeString(fmt.Sprintf(",\n    \"status\": \"%s\"", state))
}

func (this *httpRequest) writeErrors() bool {
	var err errors.Error
	ok := true
loop:
	for ok {
		select {
		case err, ok = <-this.Errors():
			if ok {
				if this.errorCount == 0 {
					this.writeString(",\n    \"errors\": [")
				}
				ok = this.writeError(err, this.errorCount)
				this.errorCount++
			}
		default:
			break loop
		}
	}

	return this.errorCount == 0 || this.writeString("\n    ]")
}

func (this *httpRequest) writeWarnings() bool {
	var err errors.Error
	ok := true
loop:
	for ok {
		select {
		case err, ok = <-this.Warnings():
			if ok {
				if this.warningCount == 0 {
					this.writeString(",\n    \"warnings\": [")
				}
				ok = this.writeError(err, this.warningCount)
				this.warningCount++
			}
		default:
			break loop
		}
	}

	return this.warningCount == 0 || this.writeString("\n    ]")
}

func (this *httpRequest) writeError(err errors.Error, count int) bool {
	var rv bool
	if count == 0 {
		rv = this.writeString("\n")
	} else {
		rv = this.writeString(",\n")
	}

	m := map[string]interface{}{
		"code": err.Code(),
		"msg":  err.Error(),
	}
	bytes, er := json.MarshalIndent(m, "        ", "    ")
	if er != nil {
		return false
	}

	return rv &&
		this.writeString("        ") &&
		this.writeString(string(bytes))
}

func (this *httpRequest) writeMetrics(metrics bool) bool {
	m := this.Metrics()
	if m == value.FALSE ||
		(m == value.NONE && !metrics) {
		return true
	}

	ts := time.Since(this.ServiceTime())
	tr := time.Since(this.RequestTime())
	rv := this.writeString(",\n    \"metrics\": {") &&
		this.writeString(fmt.Sprintf("\n        \"elapsedTime\": \"%v\"", tr)) &&
		this.writeString(fmt.Sprintf(",\n        \"executionTime\": \"%v\"", ts)) &&
		this.writeString(fmt.Sprintf(",\n        \"resultCount\": %d", this.resultCount)) &&
		this.writeString(fmt.Sprintf(",\n        \"resultSize\": %d", this.resultSize))

    phaseTimes := this.PhaseTimes()
	for k, v := range phaseTimes {
        this.writeString(fmt.Sprintf(",\n        \"%s\": \"%v\"", k, v))
	}

    var stats debug.GCStats
    debug.ReadGCStats(&stats)

    if stats.LastGC.After(this.RequestTime()) {
        this.writeString(fmt.Sprintf(",\n        \"garbageCollection\": true"))
    } else {
        this.writeString(fmt.Sprintf(",\n        \"garbageCollection\": false"))
    }

	if this.MutationCount() > 0 {
		rv = rv && this.writeString(fmt.Sprintf(",\n        \"mutationCount\": %d", this.MutationCount()))
	}

	if this.SortCount() > 0 {
		rv = rv && this.writeString(fmt.Sprintf(",\n        \"sortCount\": %d", this.SortCount()))
	}

	if this.errorCount > 0 {
		rv = rv && this.writeString(fmt.Sprintf(",\n        \"errorCount\": %d", this.errorCount))
	}

	if this.warningCount > 0 {
		rv = rv && this.writeString(fmt.Sprintf(",\n        \"warningCount\": %d", this.warningCount))
	}

	return rv && this.writeString("\n    }")
}

// responseDataManager is an interface for managing response data. It is used by httpRequest to take care of
// the data in a response.
type responseDataManager interface {
	writeString(string) bool // write the given string for the response
	noMoreData()             // action to take when there is no more data for the response
}

// bufferedWriter is an implementation of responseDataManager that writes response data to a buffer,
// up to a threshold:
type bufferedWriter struct {
	sync.Mutex
	req         *httpRequest  // the request for the response we are writing
	buffer      *bytes.Buffer // buffer for writing response data to
	buffer_pool BufferPool    // buffer manager for our buffer
	closed      bool
}

func NewBufferedWriter(r *httpRequest, bp BufferPool) *bufferedWriter {
	return &bufferedWriter{
		req:         r,
		buffer:      bp.GetBuffer(),
		buffer_pool: bp,
		closed:      false,
	}
}

func (this *bufferedWriter) writeString(s string) bool {
	this.Lock()
	defer this.Unlock()

	if this.closed {
		return false
	}

	if len(s)+len(this.buffer.Bytes()) > this.buffer_pool.BufferCapacity() { // threshold exceeded
		w := this.req.resp // our request's response writer
		// write response header and data buffered so far using request's response writer:
		w.WriteHeader(this.req.httpCode())
		io.Copy(w, this.buffer)
		// switch to non-buffered mode; change our request's responseDataManager to be a directWriter:
		this.req.writer = NewDirectWriter(this.req)
		// return buffer to pool, because response data will be directly written from now:
		this.buffer_pool.PutBuffer(this.buffer)
		this.closed = true
		// write out the string - using just-created directWriter:
		return this.req.writer.writeString(s)
	}
	// under threshold - write the string to our buffer
	_, err := this.buffer.Write([]byte(s))
	return err == nil
}

func (this *bufferedWriter) noMoreData() {
	this.Lock()
	defer this.Unlock()

	if this.closed {
		return
	}

	w := this.req.resp // our request's response writer
	r := this.req.req  // our request's http request
	// calculate and set the Content-Length header:
	content_len := strconv.Itoa(len(this.buffer.Bytes()))
	w.Header().Set("Content-Length", content_len)
	// write response header and data buffered so far:
	w.WriteHeader(this.req.httpCode())
	io.Copy(w, this.buffer)
	// no more data in the response => return buffer to pool:
	this.buffer_pool.PutBuffer(this.buffer)
	r.Body.Close()
	this.closed = true
}

// directWriter is an implementation of responseDataManager that uses the request's
// response writer to write out the data for a response
type directWriter struct {
	sync.Mutex
	req    *httpRequest // the request for the response we are writing
	closed bool
}

func NewDirectWriter(r *httpRequest) *directWriter {
	return &directWriter{
		req:    r,
		closed: false,
	}
}

// write and flush the given string using our request's response writer:
func (this *directWriter) writeString(s string) bool {
	this.Lock()
	defer this.Unlock()

	if this.closed {
		return false
	}
	w := this.req.resp
	_, err := io.WriteString(w, s)
	w.(http.Flusher).Flush()
	return err == nil
}

func (this *directWriter) noMoreData() {
	this.Lock()
	defer this.Unlock()

	if this.closed {
		return
	}
	r := this.req.req // our request's http request
	r.Body.Close()
	this.closed = true
}
