package ecsstate

// Local representation of an ECS cluster and stored by gorm
type Cluster struct {
	ARN                string `sql:"size:1024" gorm:"primary_key"`
	Name               string `sql:"unique"`
	Status             string
	ContainerInstances []ContainerInstance
	Tasks              []Task
}
