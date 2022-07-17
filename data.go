package skipper

type Data map[string]interface{}

func (d Data) HasKey(k string) bool {
	if _, ok := d[k]; ok {
		return true
	}
	return false
}

func (d Data) Get(k string) Data {
	return d[k].(Data)
}

func mergeData(a, b Data) Data {
	out := make(Data, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(Data); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(Data); ok {
					out[k] = mergeData(bv, v)
					continue
				}
			}
		}
		if v, ok := v.([]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.([]interface{}); ok {
					out[k] = append(bv, v...)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}
