package bw2bind

import "time"

type MetadataTuple struct {
	Value     string `msgpack:"val"`
	Timestamp int64  `msgpack:"ts"`
}

func (m *MetadataTuple) Time() time.Time {
	return time.Unix(0, m.Timestamp)
}

//StringPayloadObject implements 64.0.1.0/32 : String
func CreateStringPayloadObject(v string) TextPayloadObject {
	return CreateTextPayloadObject(FromDotForm("64.0.1.0"), v)
}
