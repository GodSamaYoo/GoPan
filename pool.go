package main

type Tasks struct {
	length int64
	email ,path1,path2,path3 string
}

func (t *Tasks)Execute()  {
	FileUpOneDrive(t.length,t.email ,t.path1,t.path2,t.path3)
}
type Pool struct {
	WorkNum int
	EnterChannel chan *Tasks
	JobsChannel chan *Tasks
}

func (p *Pool)worker()  {
	for {
		a := <-p.JobsChannel
		a.Execute()
	}
}

func (p *Pool)Run()  {
	for i:=0;i<p.WorkNum;i++ {
		go p.worker()
	}
	for task := range p.EnterChannel {
		p.JobsChannel <- task
	}
	close(p.JobsChannel)
	close(p.EnterChannel)
}
func NewPool(num int) *Pool {
	p := Pool{
		WorkNum:     num,
		EnterChannel: make(chan *Tasks,500),
		JobsChannel: make(chan *Tasks),
	}
	return &p
}
