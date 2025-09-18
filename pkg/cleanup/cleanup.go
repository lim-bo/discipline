package cleanup

import "log"

type Job struct {
	Name string
	F    func() error
}

var (
	jobs []*Job
)

func Register(j *Job) {
	jobs = append(jobs, j)
}

func CleanUp() {
	for _, j := range jobs {
		log.Printf("Cleanup job %s started...", j.Name)
		err := j.F()
		if err != nil {
			log.Printf("Job finished with error: %v", err)
		} else {
			log.Println("Cleaned")
		}
	}
}
