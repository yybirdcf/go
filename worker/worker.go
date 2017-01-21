package worker

type Worker struct {
	pool    chan chan Job
	jobChan chan Job
	quit    chan bool
}

func (w *Worker) Run() {
	go func() {
		for {
			//向连接池注册worker工作队列
			w.pool <- w.jobChan

			select {
			case job := <-w.jobChan:
				job.Run()
			case <-w.quit:
				return
			}
		}
	}()
}

func NewWorker(pool chan chan Job) *Worker {
	return &Worker{
		pool:    pool,
		jobChan: make(chan Job),
		quit:    make(chan bool),
	}
}
