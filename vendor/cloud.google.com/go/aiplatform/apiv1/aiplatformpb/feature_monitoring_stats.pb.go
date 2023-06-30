// Copyright 2023 Google LLC
//
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

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v3.21.12
// source: google/cloud/aiplatform/v1/feature_monitoring_stats.proto

package aiplatformpb

import (
	reflect "reflect"
	sync "sync"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Stats and Anomaly generated at specific timestamp for specific Feature.
// The start_time and end_time are used to define the time range of the dataset
// that current stats belongs to, e.g. prediction traffic is bucketed into
// prediction datasets by time window. If the Dataset is not defined by time
// window, start_time = end_time. Timestamp of the stats and anomalies always
// refers to end_time. Raw stats and anomalies are stored in stats_uri or
// anomaly_uri in the tensorflow defined protos. Field data_stats contains
// almost identical information with the raw stats in Vertex AI
// defined proto, for UI to display.
type FeatureStatsAnomaly struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Feature importance score, only populated when cross-feature monitoring is
	// enabled. For now only used to represent feature attribution score within
	// range [0, 1] for
	// [ModelDeploymentMonitoringObjectiveType.FEATURE_ATTRIBUTION_SKEW][google.cloud.aiplatform.v1.ModelDeploymentMonitoringObjectiveType.FEATURE_ATTRIBUTION_SKEW]
	// and
	// [ModelDeploymentMonitoringObjectiveType.FEATURE_ATTRIBUTION_DRIFT][google.cloud.aiplatform.v1.ModelDeploymentMonitoringObjectiveType.FEATURE_ATTRIBUTION_DRIFT].
	Score float64 `protobuf:"fixed64,1,opt,name=score,proto3" json:"score,omitempty"`
	// Path of the stats file for current feature values in Cloud Storage bucket.
	// Format: gs://<bucket_name>/<object_name>/stats.
	// Example: gs://monitoring_bucket/feature_name/stats.
	// Stats are stored as binary format with Protobuf message
	// [tensorflow.metadata.v0.FeatureNameStatistics](https://github.com/tensorflow/metadata/blob/master/tensorflow_metadata/proto/v0/statistics.proto).
	StatsUri string `protobuf:"bytes,3,opt,name=stats_uri,json=statsUri,proto3" json:"stats_uri,omitempty"`
	// Path of the anomaly file for current feature values in Cloud Storage
	// bucket.
	// Format: gs://<bucket_name>/<object_name>/anomalies.
	// Example: gs://monitoring_bucket/feature_name/anomalies.
	// Stats are stored as binary format with Protobuf message
	// Anoamlies are stored as binary format with Protobuf message
	// [tensorflow.metadata.v0.AnomalyInfo]
	// (https://github.com/tensorflow/metadata/blob/master/tensorflow_metadata/proto/v0/anomalies.proto).
	AnomalyUri string `protobuf:"bytes,4,opt,name=anomaly_uri,json=anomalyUri,proto3" json:"anomaly_uri,omitempty"`
	// Deviation from the current stats to baseline stats.
	//   1. For categorical feature, the distribution distance is calculated by
	//      L-inifinity norm.
	//   2. For numerical feature, the distribution distance is calculated by
	//      Jensen–Shannon divergence.
	DistributionDeviation float64 `protobuf:"fixed64,5,opt,name=distribution_deviation,json=distributionDeviation,proto3" json:"distribution_deviation,omitempty"`
	// This is the threshold used when detecting anomalies.
	// The threshold can be changed by user, so this one might be different from
	// [ThresholdConfig.value][google.cloud.aiplatform.v1.ThresholdConfig.value].
	AnomalyDetectionThreshold float64 `protobuf:"fixed64,9,opt,name=anomaly_detection_threshold,json=anomalyDetectionThreshold,proto3" json:"anomaly_detection_threshold,omitempty"`
	// The start timestamp of window where stats were generated.
	// For objectives where time window doesn't make sense (e.g. Featurestore
	// Snapshot Monitoring), start_time is only used to indicate the monitoring
	// intervals, so it always equals to (end_time - monitoring_interval).
	StartTime *timestamppb.Timestamp `protobuf:"bytes,7,opt,name=start_time,json=startTime,proto3" json:"start_time,omitempty"`
	// The end timestamp of window where stats were generated.
	// For objectives where time window doesn't make sense (e.g. Featurestore
	// Snapshot Monitoring), end_time indicates the timestamp of the data used to
	// generate stats (e.g. timestamp we take snapshots for feature values).
	EndTime *timestamppb.Timestamp `protobuf:"bytes,8,opt,name=end_time,json=endTime,proto3" json:"end_time,omitempty"`
}

func (x *FeatureStatsAnomaly) Reset() {
	*x = FeatureStatsAnomaly{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_aiplatform_v1_feature_monitoring_stats_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FeatureStatsAnomaly) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FeatureStatsAnomaly) ProtoMessage() {}

func (x *FeatureStatsAnomaly) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_aiplatform_v1_feature_monitoring_stats_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FeatureStatsAnomaly.ProtoReflect.Descriptor instead.
func (*FeatureStatsAnomaly) Descriptor() ([]byte, []int) {
	return file_google_cloud_aiplatform_v1_feature_monitoring_stats_proto_rawDescGZIP(), []int{0}
}

func (x *FeatureStatsAnomaly) GetScore() float64 {
	if x != nil {
		return x.Score
	}
	return 0
}

func (x *FeatureStatsAnomaly) GetStatsUri() string {
	if x != nil {
		return x.StatsUri
	}
	return ""
}

func (x *FeatureStatsAnomaly) GetAnomalyUri() string {
	if x != nil {
		return x.AnomalyUri
	}
	return ""
}

func (x *FeatureStatsAnomaly) GetDistributionDeviation() float64 {
	if x != nil {
		return x.DistributionDeviation
	}
	return 0
}

func (x *FeatureStatsAnomaly) GetAnomalyDetectionThreshold() float64 {
	if x != nil {
		return x.AnomalyDetectionThreshold
	}
	return 0
}

func (x *FeatureStatsAnomaly) GetStartTime() *timestamppb.Timestamp {
	if x != nil {
		return x.StartTime
	}
	return nil
}

func (x *FeatureStatsAnomaly) GetEndTime() *timestamppb.Timestamp {
	if x != nil {
		return x.EndTime
	}
	return nil
}

var File_google_cloud_aiplatform_v1_feature_monitoring_stats_proto protoreflect.FileDescriptor

var file_google_cloud_aiplatform_v1_feature_monitoring_stats_proto_rawDesc = []byte{
	0x0a, 0x39, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2f, 0x61,
	0x69, 0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x2f, 0x76, 0x31, 0x2f, 0x66, 0x65, 0x61,
	0x74, 0x75, 0x72, 0x65, 0x5f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x5f,
	0x73, 0x74, 0x61, 0x74, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1a, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x69, 0x70, 0x6c, 0x61, 0x74,
	0x66, 0x6f, 0x72, 0x6d, 0x2e, 0x76, 0x31, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61,
	0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xd2, 0x02, 0x0a, 0x13, 0x46, 0x65, 0x61,
	0x74, 0x75, 0x72, 0x65, 0x53, 0x74, 0x61, 0x74, 0x73, 0x41, 0x6e, 0x6f, 0x6d, 0x61, 0x6c, 0x79,
	0x12, 0x14, 0x0a, 0x05, 0x73, 0x63, 0x6f, 0x72, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x01, 0x52,
	0x05, 0x73, 0x63, 0x6f, 0x72, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x73, 0x74, 0x61, 0x74, 0x73, 0x5f,
	0x75, 0x72, 0x69, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x73, 0x74, 0x61, 0x74, 0x73,
	0x55, 0x72, 0x69, 0x12, 0x1f, 0x0a, 0x0b, 0x61, 0x6e, 0x6f, 0x6d, 0x61, 0x6c, 0x79, 0x5f, 0x75,
	0x72, 0x69, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x61, 0x6e, 0x6f, 0x6d, 0x61, 0x6c,
	0x79, 0x55, 0x72, 0x69, 0x12, 0x35, 0x0a, 0x16, 0x64, 0x69, 0x73, 0x74, 0x72, 0x69, 0x62, 0x75,
	0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x64, 0x65, 0x76, 0x69, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x01, 0x52, 0x15, 0x64, 0x69, 0x73, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x69,
	0x6f, 0x6e, 0x44, 0x65, 0x76, 0x69, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x3e, 0x0a, 0x1b, 0x61,
	0x6e, 0x6f, 0x6d, 0x61, 0x6c, 0x79, 0x5f, 0x64, 0x65, 0x74, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x5f, 0x74, 0x68, 0x72, 0x65, 0x73, 0x68, 0x6f, 0x6c, 0x64, 0x18, 0x09, 0x20, 0x01, 0x28, 0x01,
	0x52, 0x19, 0x61, 0x6e, 0x6f, 0x6d, 0x61, 0x6c, 0x79, 0x44, 0x65, 0x74, 0x65, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x54, 0x68, 0x72, 0x65, 0x73, 0x68, 0x6f, 0x6c, 0x64, 0x12, 0x39, 0x0a, 0x0a, 0x73,
	0x74, 0x61, 0x72, 0x74, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x73, 0x74, 0x61,
	0x72, 0x74, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x35, 0x0a, 0x08, 0x65, 0x6e, 0x64, 0x5f, 0x74, 0x69,
	0x6d, 0x65, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x52, 0x07, 0x65, 0x6e, 0x64, 0x54, 0x69, 0x6d, 0x65, 0x42, 0xd9, 0x01,
	0x0a, 0x1e, 0x63, 0x6f, 0x6d, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f,
	0x75, 0x64, 0x2e, 0x61, 0x69, 0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x2e, 0x76, 0x31,
	0x42, 0x1b, 0x46, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x4d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72,
	0x69, 0x6e, 0x67, 0x53, 0x74, 0x61, 0x74, 0x73, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a,
	0x3e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x67, 0x6f, 0x2f, 0x61, 0x69, 0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x2f,
	0x61, 0x70, 0x69, 0x76, 0x31, 0x2f, 0x61, 0x69, 0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d,
	0x70, 0x62, 0x3b, 0x61, 0x69, 0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x70, 0x62, 0xaa,
	0x02, 0x1a, 0x47, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x43, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x41,
	0x49, 0x50, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x1a, 0x47,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x5c, 0x43, 0x6c, 0x6f, 0x75, 0x64, 0x5c, 0x41, 0x49, 0x50, 0x6c,
	0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x5c, 0x56, 0x31, 0xea, 0x02, 0x1d, 0x47, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x3a, 0x3a, 0x43, 0x6c, 0x6f, 0x75, 0x64, 0x3a, 0x3a, 0x41, 0x49, 0x50, 0x6c, 0x61,
	0x74, 0x66, 0x6f, 0x72, 0x6d, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_google_cloud_aiplatform_v1_feature_monitoring_stats_proto_rawDescOnce sync.Once
	file_google_cloud_aiplatform_v1_feature_monitoring_stats_proto_rawDescData = file_google_cloud_aiplatform_v1_feature_monitoring_stats_proto_rawDesc
)

func file_google_cloud_aiplatform_v1_feature_monitoring_stats_proto_rawDescGZIP() []byte {
	file_google_cloud_aiplatform_v1_feature_monitoring_stats_proto_rawDescOnce.Do(func() {
		file_google_cloud_aiplatform_v1_feature_monitoring_stats_proto_rawDescData = protoimpl.X.CompressGZIP(file_google_cloud_aiplatform_v1_feature_monitoring_stats_proto_rawDescData)
	})
	return file_google_cloud_aiplatform_v1_feature_monitoring_stats_proto_rawDescData
}

var file_google_cloud_aiplatform_v1_feature_monitoring_stats_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_google_cloud_aiplatform_v1_feature_monitoring_stats_proto_goTypes = []interface{}{
	(*FeatureStatsAnomaly)(nil),   // 0: google.cloud.aiplatform.v1.FeatureStatsAnomaly
	(*timestamppb.Timestamp)(nil), // 1: google.protobuf.Timestamp
}
var file_google_cloud_aiplatform_v1_feature_monitoring_stats_proto_depIdxs = []int32{
	1, // 0: google.cloud.aiplatform.v1.FeatureStatsAnomaly.start_time:type_name -> google.protobuf.Timestamp
	1, // 1: google.cloud.aiplatform.v1.FeatureStatsAnomaly.end_time:type_name -> google.protobuf.Timestamp
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_google_cloud_aiplatform_v1_feature_monitoring_stats_proto_init() }
func file_google_cloud_aiplatform_v1_feature_monitoring_stats_proto_init() {
	if File_google_cloud_aiplatform_v1_feature_monitoring_stats_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_google_cloud_aiplatform_v1_feature_monitoring_stats_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FeatureStatsAnomaly); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_google_cloud_aiplatform_v1_feature_monitoring_stats_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_google_cloud_aiplatform_v1_feature_monitoring_stats_proto_goTypes,
		DependencyIndexes: file_google_cloud_aiplatform_v1_feature_monitoring_stats_proto_depIdxs,
		MessageInfos:      file_google_cloud_aiplatform_v1_feature_monitoring_stats_proto_msgTypes,
	}.Build()
	File_google_cloud_aiplatform_v1_feature_monitoring_stats_proto = out.File
	file_google_cloud_aiplatform_v1_feature_monitoring_stats_proto_rawDesc = nil
	file_google_cloud_aiplatform_v1_feature_monitoring_stats_proto_goTypes = nil
	file_google_cloud_aiplatform_v1_feature_monitoring_stats_proto_depIdxs = nil
}
