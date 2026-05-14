// Package feed connects to the FeedConstruct RMQ source (or replays raw/json
// samples) and persists every delivery into raw_messages before fan-out.
//
// Maps to upload-guideline 业务域 "连接接入" + "原始消息" (M01).
// BDD scaffold — see *_test.go in this package.
package feed
