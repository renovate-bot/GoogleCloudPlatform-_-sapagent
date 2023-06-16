/*
Copyright 2023 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package workloadmanager

import (
	"context"
	"testing"

	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoredresourcepb "google.golang.org/genproto/googleapis/api/monitoredres"
	cpb "google.golang.org/genproto/googleapis/monitoring/v3"
	mrpb "google.golang.org/genproto/googleapis/monitoring/v3"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	"github.com/GoogleCloudPlatform/sapagent/internal/cloudmonitoring/fake"
	"github.com/GoogleCloudPlatform/sapagent/internal/hanainsights/ruleengine"
)

func TestProcessInsights(t *testing.T) {
	params := Parameters{
		Config:            defaultConfigurationDBMetrics,
		TimeSeriesCreator: &fake.TimeSeriesCreator{},
		BackOffs:          defaultBackOffIntervals,
	}

	insights := make(ruleengine.Insights)
	insights["rule_id"] = []ruleengine.ValidationResult{
		ruleengine.ValidationResult{
			RecommendationID: "recommendation_1",
			Result:           true,
		},
		ruleengine.ValidationResult{
			RecommendationID: "recommendation_2",
		},
	}

	want := WorkloadMetrics{
		Metrics: []*mrpb.TimeSeries{{
			Metric: &metricpb.Metric{
				Type: "workload.googleapis.com/sap/validation/hanasecurity",
				Labels: map[string]string{
					"rule_id_recommendation_1": "false",
					"rule_id_recommendation_2": "true",
				},
			},
			MetricKind: metricpb.MetricDescriptor_GAUGE,
			Resource: &monitoredresourcepb.MonitoredResource{
				Type: "gce_instance",
				Labels: map[string]string{
					"instance_id": "test-instance-id",
					"zone":        "test-zone",
					"project_id":  "test-project-id",
				},
			},
			Points: []*mrpb.Point{{
				Interval: &cpb.TimeInterval{},
				Value: &cpb.TypedValue{
					Value: &cpb.TypedValue_DoubleValue{
						DoubleValue: 1,
					},
				},
			}},
		}},
	}

	got := processInsights(context.Background(), params, insights)
	if diff := cmp.Diff(want, got, protocmp.Transform(), protocmp.IgnoreFields(&cpb.TimeInterval{}, "start_time", "end_time")); diff != "" {
		t.Errorf("processInsights() failure diff (-want +got):\n%s", diff)
	}
}
