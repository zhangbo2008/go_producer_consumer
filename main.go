/*如何实现百万并发


http://marcio.io/2015/07/handling-1-million-requests-per-minute-with-golang/

总结,圣餐这,消费者模型.就是利用channel把他们都链接起来.  所有的channel都是有maxsize来设置大小的.防止资源爆炸


首先需要一个派遣器dispatcher.
在派遣器里面设置一个jobs:chan chan job.用来注册所有的job.
然后建立一堆worker,每一个worker都需要共享这个变量jobs来建立.这样每一个worker就共享了全部的任务.


先启动一堆消费者,让他们每一个都有自己的channel来存放东西,然后他们不停的那东西,处理掉


再启动圣餐这. 不停的从channel1里面把东西放到消费者的chanenl里面.


再启动接口, 接口每次盗用把东西放到channel1里面.
这样就都跑起来了. 只是这个项目不知道哪里有bug,没跑起来.


*/
package main

// Job represents the job to be run
type Job struct {
	Payload Payload
}

// A buffered channel that we can send work requests on.
var JobQueue chan Job

// Worker represents the worker that executes the job
type Worker struct {
	WorkerPool  chan chan Job
	JobChannel  chan Job
	quit    	chan bool
}

func NewWorker(workerPool chan chan Job) Worker {
	return Worker{
		WorkerPool: workerPool,
		JobChannel: make(chan Job),
		quit:       make(chan bool)}
}



func payloadHandler() {

print("running")

	//// Read the body into a string for json decoding
	var content = &PayloadCollection{Payloads:   []Payload{{"111"},{"2222"},{"33333"}}}
	//err := json.NewDecoder(io.LimitReader(r.Body, MaxLength)).Decode(&content)
	//if err != nil {
	//	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	//	w.WriteHeader(http.StatusBadRequest)
	//	return
	//}

	// Go through each payload and queue items individually to be posted to S3
	for _, payload := range content.Payloads {

		// let's create a job with the payload
		work := Job{Payload: payload}  //简历bob

		// Push the work onto the queue.
		JobQueue <- work   //job放入jobqueue中.
	}

	return
}




type PayloadCollection struct {//  处理的任务.
	WindowsVersion  string    `json:"version"`
	Token           string    `json:"token"`
	Payloads        []Payload `json:"data"`
}

type Payload struct {
	// [redacted]
	token string
}








func (p *Payload) UploadToS3() error {//任务调用函数
	print("running serve!!!!!!!!!!!!!!!")
	return nil
}










// Start method starts the run loop for the worker, listening for a quit channel in
// case we need to stop it

// worker 就是一个消费者,不停的从channel中取物体.
func (w Worker) Start() {
	go func() {
		for {
			// register the current worker into the worker queue.
			w.WorkerPool <- w.JobChannel

			select {
			case job := <-w.JobChannel:
				// we have received a work request.
				if err := job.Payload.UploadToS3(); err != nil {
					print("log")
				}

			case <-w.quit:
				// we have received a signal to stop
				return
			}
		}
	}()
}
//调用stop就会停止.
// Stop signals the worker to stop listening for work requests.
func (w Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}
type Dispatcher struct {
	// A pool of workers channels that are registered with the dispatcher
	WorkerPool chan chan Job
	maxWorkers int
	pool int
}

func NewDispatcher(maxWorkers int) *Dispatcher {
	pool := make(chan chan Job, maxWorkers)
	return &Dispatcher{WorkerPool: pool,	maxWorkers:100	,	pool:100}
}

func (d *Dispatcher) Run() {
	// starting n number of workers           //先启动消费者!!!!!!!!!!!
	for i := 0; i < d.maxWorkers; i++ {
		worker := NewWorker(d.WorkerPool)
		worker.Start()
	}
// 再启动圣餐这
	go d.dispatch()
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case job := <-JobQueue:
			// a job request has been received
			go func(job Job) {
				// try to obtain a worker job channel that is available.
				// this will block until a worker is idle
				jobChannel := <-d.WorkerPool

				// dispatch the job to the worker job channel
				jobChannel <- job
			}(job)
		}
	}
}










func main(){

	MaxWorker:=100
	//MaxQueue:=100 //dispatcher 整合了work, job ,queue
	dispatcher := NewDispatcher(MaxWorker)


	go dispatcher.Run()
//go func() {for { ;print("111111");}}()
	go payloadHandler()
	for { ;;}

	//下面我们就疯狂调用 payloadHandler 这个函数即可.开启全部流程.
}