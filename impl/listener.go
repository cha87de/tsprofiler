package impl

func (profiler *simpleProfiler) listener() {
	itemCount := 0
	for {
		input := <-profiler.in
		profiler.dataaccess.Lock()
		profiler.cpudata = append(profiler.cpudata, input)
		profiler.dataaccess.Unlock()
		itemCount++
		if itemCount > profiler.settings.BufferSize {
			// buffer is full, trigger profiler
			go profiler.profile()
			itemCount = 0
		}
	}
}
