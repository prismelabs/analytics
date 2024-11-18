package sessionstorage

type gcJob struct {
	jobIndex int
	deviceId uint64
	pExpiry  uint32
}

// gcQueue is a non thread-safe priority queue.
type gcQueue []*gcJob

// Len implements heap.Interface.
func (gq *gcQueue) Len() int {
	return len(*gq)
}

// Less implements heap.Interface.
func (gq *gcQueue) Less(i int, j int) bool {
	jobs := *gq
	return jobs[i].pExpiry < jobs[j].pExpiry
}

// Pop implements heap.Interface.
func (gq *gcQueue) Pop() any {
	jobs := *gq
	n := len(jobs)
	j := jobs[n-1]
	j.jobIndex = -1
	jobs[n-1] = nil
	*gq = jobs[0 : n-1]
	return j
}

// Push implements heap.Interface.
func (gq *gcQueue) Push(x any) {
	jobs := *gq
	j := x.(*gcJob)
	j.jobIndex = len(jobs)
	jobs = append(jobs, j)
	*gq = jobs
}

// Swap implements heap.Interface.
func (gq *gcQueue) Swap(i int, j int) {
	jobs := *gq
	jobs[i], jobs[j] = jobs[j], jobs[i]
	jobs[i].jobIndex = i
	jobs[j].jobIndex = j
	*gq = jobs
}
