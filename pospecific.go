package bw2bind

import "time"

type MetadataTuple struct {
	Value     string `msgpack:"val"`
	Timestamp int64  `msgpack:"ts"`
}

func (m *MetadataTuple) Time() time.Time {
	return time.Unix(0, m.Timestamp)
}
