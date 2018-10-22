package impl

func (profiler *simpleProfiler) listener() {
	for {
		input := <-profiler.in
		profiler.dataaccess.Lock()
		profiler.data = append(profiler.data, input)
		profiler.dataaccess.Unlock()
	}
}
