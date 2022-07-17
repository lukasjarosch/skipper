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

func (d Data) MergeReplace(data Data) Data {
	out := make(Data, len(d))
	for k, v := range d {
		out[k] = v
	}
	for k, v := range data {
		if v, ok := v.(Data); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(Data); ok {
					out[k] = bv.MergeReplace(v)
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
