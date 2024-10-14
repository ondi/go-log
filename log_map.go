//
//
//

package log

type Queue_map_t map[string]Queue

type Level_map_t map[int64]Queue_map_t

func NewLogMap() (self Level_map_t) {
	self = Level_map_t{}
	return
}

func (self Level_map_t) AddOutput(level_id int64, writer_name string, queue Queue) (ok bool) {
	writers, ok := self[level_id]
	if !ok {
		writers = Queue_map_t{}
		self[level_id] = writers
	}
	_, ok = writers[writer_name]
	if !ok {
		writers[writer_name] = queue
		return true
	}
	return false
}

func (self Level_map_t) DelOutput(level_id int64, writer_name string) (ok bool) {
	writers, ok := self[level_id]
	if !ok {
		return
	}
	writer, ok := writers[writer_name]
	if !ok {
		return
	}
	writer.Close()
	delete(writers, writer_name)
	return
}

func (self Level_map_t) AddOutputs(writer_name string, queue Queue, levels []Info_t) (ok bool) {
	for _, v := range levels {
		if ok = self.AddOutput(v.LevelId, writer_name, queue); !ok {
			return
		}
	}
	return
}

func (self Level_map_t) DelOutputs(writer_name string, levels []Info_t) (ok bool) {
	for _, v := range levels {
		if ok = self.DelOutput(v.LevelId, writer_name); !ok {
			return
		}
	}
	return
}

func CopyLevelMap(in Level_map_t) (out Level_map_t) {
	var ok bool
	var temp Queue_map_t
	out = Level_map_t{}
	for k1, v1 := range in {
		for k2, v2 := range v1 {
			if temp, ok = out[k1]; !ok {
				temp = Queue_map_t{}
				out[k1] = temp
			}
			temp[k2] = v2
		}
	}
	return
}
