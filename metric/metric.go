package metric

import (
	"fmt"
	"hash/fnv"
	"sort"
	"time"

	"github.com/eviltomorrow/omega"
)

type metric struct {
	name   string
	tags   []*omega.Tag
	fields []*omega.Field
	tm     time.Time

	tp omega.ValueType
}

func New(name string, tags map[string]string, fields map[string]interface{}, tm time.Time, tp ...omega.ValueType) omega.Metric {
	var vtype omega.ValueType
	if len(tp) > 0 {
		vtype = tp[0]
	} else {
		vtype = omega.Unknown
	}

	m := &metric{
		name:   name,
		tags:   nil,
		fields: nil,
		tm:     tm,
		tp:     vtype,
	}

	if len(tags) > 0 {
		m.tags = make([]*omega.Tag, 0, len(tags))
		for k, v := range tags {
			m.tags = append(m.tags,
				&omega.Tag{Key: k, Value: v})
		}
		sort.Slice(m.tags, func(i, j int) bool { return m.tags[i].Key < m.tags[j].Key })
	}

	if len(fields) > 0 {
		m.fields = make([]*omega.Field, 0, len(fields))
		for k, v := range fields {
			v := convertField(v)
			if v == nil {
				continue
			}
			m.AddField(k, v)
		}
	}

	return m
}

func (m *metric) Fields() map[string]interface{} {
	fields := make(map[string]interface{}, len(m.fields))
	for _, field := range m.fields {
		fields[field.Key] = field.Value
	}
	return fields
}

func (m *metric) AddField(key string, value interface{}) {
	for i, field := range m.fields {
		if key == field.Key {
			m.fields[i] = &omega.Field{Key: key, Value: convertField(value)}
			return
		}
	}
	m.fields = append(m.fields, &omega.Field{Key: key, Value: convertField(value)})
}

func (m *metric) Name() string {
	return m.name
}

func (m *metric) SetName(name string) {
	m.name = name
}

func (m *metric) Tags() map[string]string {
	tags := make(map[string]string, len(m.tags))
	for _, tag := range m.tags {
		tags[tag.Key] = tag.Value
	}
	return tags
}

func (m *metric) AddTag(key, value string) {
	var tag = &omega.Tag{Key: key, Value: value}
	m.tags = append(m.tags, tag)
}

func (m *metric) Time() time.Time {
	return m.tm
}

func (m *metric) SetTime(t time.Time) {
	m.tm = t
}

func (m *metric) Type() omega.ValueType {
	return m.tp
}

func (m *metric) Copy() omega.Metric {
	m2 := &metric{
		name:   m.name,
		tags:   make([]*omega.Tag, len(m.tags)),
		fields: make([]*omega.Field, len(m.fields)),
		tm:     m.tm,
		tp:     m.tp,
	}

	for i, tag := range m.tags {
		m2.tags[i] = &omega.Tag{Key: tag.Key, Value: tag.Value}
	}

	for i, field := range m.fields {
		m2.fields[i] = &omega.Field{Key: field.Key, Value: field.Value}
	}
	return m2
}

func (m *metric) HashID() uint64 {
	h := fnv.New64a()
	h.Write([]byte(m.name))
	h.Write([]byte("\n"))
	for _, tag := range m.tags {
		h.Write([]byte(tag.Key))
		h.Write([]byte("\n"))
		h.Write([]byte(tag.Value))
		h.Write([]byte("\n"))
	}
	return h.Sum64()
}

func (m *metric) String() string {
	return fmt.Sprintf("%s %v %v %d", m.name, m.Tags(), m.Fields(), m.tm.UnixNano())
}

func convertField(v interface{}) interface{} {
	switch v := v.(type) {
	case float64:
		return v
	case int64:
		return v
	case string:
		return v
	case bool:
		return v
	case int:
		return int64(v)
	case uint:
		return uint64(v)
	case uint64:
		return v
	case []byte:
		return string(v)
	case int32:
		return int64(v)
	case int16:
		return int64(v)
	case int8:
		return int64(v)
	case uint32:
		return uint64(v)
	case uint16:
		return uint64(v)
	case uint8:
		return uint64(v)
	case float32:
		return float64(v)
	case *float64:
		if v != nil {
			return *v
		}
	case *int64:
		if v != nil {
			return *v
		}
	case *string:
		if v != nil {
			return *v
		}
	case *bool:
		if v != nil {
			return *v
		}
	case *int:
		if v != nil {
			return int64(*v)
		}
	case *uint:
		if v != nil {
			return uint64(*v)
		}
	case *uint64:
		if v != nil {
			return *v
		}
	case *[]byte:
		if v != nil {
			return string(*v)
		}
	case *int32:
		if v != nil {
			return int64(*v)
		}
	case *int16:
		if v != nil {
			return int64(*v)
		}
	case *int8:
		if v != nil {
			return int64(*v)
		}
	case *uint32:
		if v != nil {
			return uint64(*v)
		}
	case *uint16:
		if v != nil {
			return uint64(*v)
		}
	case *uint8:
		if v != nil {
			return uint64(*v)
		}
	case *float32:
		if v != nil {
			return float64(*v)
		}
	default:
		return nil
	}
	return nil
}
