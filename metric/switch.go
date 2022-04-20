package metric

import (
	"fmt"

	"github.com/eviltomorrow/omega"
	"github.com/eviltomorrow/omega/internal/api/collector/pb"
	"github.com/eviltomorrow/omega/pkg/zlog"
	"go.uber.org/zap"
)

func SwitchMetricsToMetricSet(metrics []omega.Metric) (*pb.MetricSet, bool) {
	if len(metrics) == 0 {
		return nil, false
	}

	var data = make([]*pb.MetricFamily, 0, len(metrics))
	for _, metric := range metrics {
		var mf = &pb.MetricFamily{
			Name: metric.Name(),
			Unit: "unknown",
			Help: "unknown",
			Tags: make([]*pb.Label, 0, len(metric.Tags())),
		}

		switch metric.Type() {
		case omega.Counter:
			mf.Fields = make([]*pb.Field, 0, len(metric.Fields()))
			mf.Type = pb.MetricType_COUNTER

			for key, val := range metric.Fields() {
				switch v := val.(type) {
				case float64:
					mf.Fields = append(mf.Fields, &pb.Field{
						Value: &pb.Field_CounterValue{
							CounterValue: &pb.CounterValue{
								Total: &pb.CounterValue_DoubleValue{
									DoubleValue: v,
								},
							},
						},
						Name: key,
					})

				case uint64:
					mf.Fields = append(mf.Fields, &pb.Field{
						Value: &pb.Field_CounterValue{
							CounterValue: &pb.CounterValue{
								Total: &pb.CounterValue_IntValue{
									IntValue: v,
								},
							},
						},
						Name: key,
					})

				default:
					zlog.Warn("assert field value failure", zap.String("key", key), zap.Any("val", val), zap.String("type", "counter"))
				}
			}

		case omega.Gauge:
			mf.Fields = make([]*pb.Field, 0, len(metric.Fields()))
			mf.Type = pb.MetricType_GAUGE

			for key, val := range metric.Fields() {
				switch v := val.(type) {
				case float64:
					mf.Fields = append(mf.Fields, &pb.Field{
						Value: &pb.Field_GaugeValue{
							GaugeValue: &pb.GaugeValue{
								Value: &pb.GaugeValue_DoubleValue{
									DoubleValue: v,
								},
							},
						},
						Name: key,
					})

				case int64:
					mf.Fields = append(mf.Fields, &pb.Field{
						Value: &pb.Field_GaugeValue{
							GaugeValue: &pb.GaugeValue{
								Value: &pb.GaugeValue_IntValue{
									IntValue: v,
								},
							},
						},
						Name: key,
					})
				case uint64:
					mf.Fields = append(mf.Fields, &pb.Field{
						Value: &pb.Field_GaugeValue{
							GaugeValue: &pb.GaugeValue{
								Value: &pb.GaugeValue_IntValue{
									IntValue: int64(v),
								},
							},
						},
						Name: key,
					})
				default:
					zlog.Warn("assert field value failure", zap.String("key", key), zap.Any("val", val), zap.String("type", fmt.Sprintf("%T", v)))
				}
			}

		case omega.StateSet:
			mf.Fields = make([]*pb.Field, 0, len(metric.Fields()))
			mf.Type = pb.MetricType_STATE_SET

			var state = make([]*pb.StateSetValue_State, 0, len(metric.Fields()))
			for key, val := range metric.Fields() {
				switch v := val.(type) {
				case bool:
					state = append(state, &pb.StateSetValue_State{
						Name:    key,
						Enabled: v,
					})

				default:
					zlog.Warn("assert field value failure", zap.String("key", key), zap.Any("val", val), zap.String("type", "stateSet"))
				}
			}
			mf.Fields = append(mf.Fields, &pb.Field{
				Value: &pb.Field_StateSetValue{
					StateSetValue: &pb.StateSetValue{
						States: state,
					},
				},
			})

		case omega.Info:
			mf.Fields = make([]*pb.Field, 0, 1)
			mf.Type = pb.MetricType_INFO

			var label = make([]*pb.Label, 0, len(metric.Fields()))
			for key, val := range metric.Fields() {
				switch v := val.(type) {
				case string:
					label = append(label, &pb.Label{
						Name:  key,
						Value: v,
					})

				default:
					zlog.Warn("assert field value failure", zap.String("key", key), zap.Any("val", val), zap.String("type", "info"))
				}
			}
			mf.Fields = append(mf.Fields, &pb.Field{
				Value: &pb.Field_InfoValue{
					InfoValue: &pb.InfoValue{
						Label: label,
					},
				},
			})

		// case omega.Summary:
		// 	mf.Type = pb.MetricType_SUMMARY
		// case omega.Histogram:
		// 	mf.Type = pb.MetricType_HISTOGRAM
		case omega.Unknown:
			mf.Type = pb.MetricType_UNKNOWN

			for key, val := range metric.Fields() {
				switch v := val.(type) {
				case float64:
					mf.Fields = append(mf.Fields, &pb.Field{
						Value: &pb.Field_UnknownValue{
							UnknownValue: &pb.UnknownValue{
								Value: &pb.UnknownValue_DoubleValue{
									DoubleValue: v,
								},
							},
						},
						Name: key,
					})

				case int64:
					mf.Fields = append(mf.Fields, &pb.Field{
						Value: &pb.Field_UnknownValue{
							UnknownValue: &pb.UnknownValue{
								Value: &pb.UnknownValue_IntValue{
									IntValue: v,
								},
							},
						},
						Name: key,
					})

				default:
					zlog.Warn("assert field value failure", zap.String("key", key), zap.Any("val", val), zap.String("type", "unknown"))
				}
			}

		default:
			zlog.Warn("not implement metric switch", zap.Any("type", metric.Type()))
			continue
		}

		for k, v := range metric.Tags() {
			mf.Tags = append(mf.Tags, &pb.Label{
				Name:  k,
				Value: v,
			})
		}
		data = append(data, mf)
	}

	return &pb.MetricSet{MetricFamilies: data}, true
}
