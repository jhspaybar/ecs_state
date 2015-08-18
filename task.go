package ecs_state

// Local representation of an ECS Task and stored by gorm.  A number of fields are absent
// for now as they are not needed to track and update the state of the state of the cluster typically.
type Task struct {
	ARN                  string `sql:"size:1024" gorm:"primary_key"`
	DesiredStatus        string
	LastStatus           string
	StartedBy            string `sql:"index"`
	ClusterARN           string `sql:"size:1024;index"`
	ContainerInstanceARN string `sql:"size:1024;index"`
	TaskDefinitionARN    string `sql:"size:1024;index"`

	// Not part of the ECS API
	RefreshTime int
}
