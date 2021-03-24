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
	JobsChannel chan *Tasks
	process int
}

func (p *Pool)worker()  {
	for a := range p.JobsChannel{
		a.Execute()
		p.process++
	}
}

func (p *Pool)Run()  {
	for i:=0;i<p.WorkNum;i++ {
		go p.worker()
	}
}
func NewPool(num int) *Pool {
	p := Pool{
		WorkNum:     num,
		JobsChannel: make(chan *Tasks),
		process: 0,
	}
	return &p
}
