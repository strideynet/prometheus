// Copyright 2016 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	"net/http"
	"time"

	"github.com/prometheus/prometheus/storage/remote"
)

// QueueRetriever provides a list of remote write queues
type QueueRetriever interface {
	Queues() []*remote.QueueManager
}

// RemoteWritesDiscovery exposes all of the configured remote writes and their status
type RemoteWritesDiscovery struct {
	Queues []*RemoteWriteQueue `json:"queues"`
}

// RemoteWrite represents a single configured remote write and its status
type RemoteWriteQueue struct {
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`

	ShardsMax     int `json:"shardsMax"`
	ShardsMin     int `json:"shardsMin"`
	ShardsCurrent int `json:"shardsCurrent"`

	ShardingCalculations RemoteWriteShardingCalculations `json:"shardingCalculations"`
	IsResharding         bool                            `json:"isResharding"`

	Shards []*RemoteWriteShard `json:"shards"`
}

// RemoteWriteShardingCalculations provides insight into the values that are used to calculate the required shard num
type RemoteWriteShardingCalculations struct {
	LastRan            time.Time `json:"lastRan"`
	Delay              float64   `json:"delay"`
	DesiredShards      float64   `json:"desiredShards"`
	HighestRecv        float64   `json:"highestRecv"`
	HighestSent        float64   `json:"highestSent"`
	SamplesInRate      float64   `json:"samplesInRate"`
	SamplesKeptRatio   float64   `json:"samplesKeptRatio"`
	SamplesOutDuration float64   `json:"samplesOutDuration"`
	SamplesOutRate     float64   `json:"samplesOutRate"`
	SamplesPending     float64   `json:"samplesPending"`
	SamplesPendingRate float64   `json:"samplesPendingRate"`
	TimePerSample      float64   `json:"timePerSample"`
}

// RemoteWriteShard represents a single shard of a remote write queue and its state
type RemoteWriteShard struct {
	PendingSamples int `json:"pendingSamples"`

	LastError        string    `json:"lastError"`
	LastSentTime     time.Time `json:"lastSentTime"`
	LastSentDuration float64   `json:"lastSentDuration"`
}

// remoteWrites handles an API request for fetching all of the configured remote writes
func (api *API) remoteWrites(r *http.Request) apiFuncResult {
	queues := api.queueRetriever(r.Context()).Queues()
	res := RemoteWritesDiscovery{
		Queues: make([]*RemoteWriteQueue, 0, len(queues)),
	}

	var wrn []error

	for _, q := range queues {
		rRQ := &RemoteWriteQueue{}
		sC := q.StoreClient()

		rRQ.Name = sC.Name()
		rRQ.Endpoint = sC.Endpoint()

		cfg := q.Config()
		rRQ.ShardsMax = cfg.MaxShards
		rRQ.ShardsMin = cfg.MinShards
		rRQ.ShardsCurrent = q.CurrentShardNum()
		rRQ.IsResharding = q.IsResharding()

		c := q.ShardingCalculations()
		rRQ.ShardingCalculations = RemoteWriteShardingCalculations{
			LastRan:            c.LastRan,
			Delay:              c.Delay,
			DesiredShards:      c.DesiredShards,
			HighestRecv:        c.HighestRecv,
			HighestSent:        c.HighestSent,
			SamplesInRate:      c.SamplesInRate,
			SamplesKeptRatio:   c.SamplesKeptRatio,
			SamplesOutDuration: c.SamplesOutDuration,
			SamplesOutRate:     c.SamplesOutRate,
			SamplesPending:     c.SamplesPending,
			SamplesPendingRate: c.SamplesPendingRate,
			TimePerSample:      c.TimePerSample,
		}

		// for _, shard := ranges rRQ.shards {
		// 		TODO: fetch various shard data.
		rRQ.Shards = append(rRQ.Shards, &RemoteWriteShard{
			PendingSamples: 1337,

			LastError:        "example err",
			LastSentTime:     time.Now(),
			LastSentDuration: 100.123,
		})
		// }

		res.Queues = append(res.Queues, rRQ)
	}

	return apiFuncResult{res, nil, wrn, nil}
}
