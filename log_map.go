//
//
//

package log

type Queue_map_t map[string]Queue

type Level_map_t map[int64]Queue_map_t

func NewLevelMap() Level_map_t {
	return Level_map_t{}
}

func (self Level_map_t) AddOutput(level_id int64, writer_name string, queue Queue) (ok bool) {
	writers, ok := self[level_id]
	if !ok {
		writers = Queue_map_t{}
		self[level_id] = writers
	}
	if _, ok = writers[writer_name]; !ok {
		writers[writer_name] = queue
		return true
	}
	return false
}

func (self Level_map_t) DelOutput(level_id int64, writer_name string) (writer Queue) {
	writers, ok := self[level_id]
	if !ok {
		return
	}
	if writer, ok = writers[writer_name]; ok {
		delete(writers, writer_name)
	}
	return
}

func (self Level_map_t) AddOutputs(writer_name string, queue Queue, levels []int64) Level_map_t {
	for _, v := range levels {
		self.AddOutput(v, writer_name, queue)
	}
	return self
}

func (self Level_map_t) DelOutputs(writer_name string, levels []int64) Level_map_t {
	for _, v := range levels {
		if writer := self.DelOutput(v, writer_name); writer != nil {
			// writer may be used at other levels
			// writer.Close()
		}
	}
	return self
}

func (self Level_map_t) Copy(out Level_map_t) Level_map_t {
	var ok bool
	var temp Queue_map_t
	for k1, v1 := range self {
		for k2, v2 := range v1 {
			if temp, ok = out[k1]; !ok {
				temp = Queue_map_t{}
				out[k1] = temp
			}
			temp[k2] = v2
		}
	}
	return out
}

func (self Level_map_t) Close() {
	writers := Queue_map_t{}
	for _, level := range self {
		for writer_name, writer := range level {
			writers[writer_name] = writer
		}
	}
	for _, v := range writers {
		v.Close()
	}
}
