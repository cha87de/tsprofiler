package impl

func (profiler *profiler) listener() {
	itemCount := 0
	for !profiler.stopped {
		input := <-profiler.input
		profiler.add(input)
		itemCount++
		if itemCount > profiler.settings.BufferSize {
			// buffer is full, trigger profiler
			go profiler.profile()
			itemCount = 0
		}
	}
}
