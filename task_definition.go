package ecs_state

// Local representation of an ECS TaskDefinition and stored by gorm.  Resources are extracted,
// but the complete definition is ignored.
type TaskDefinition struct {
	ARN         string `sql:"size:1024" gorm:"primary_key"`
	ShortString string `sql:"unique"`
	Cpu         int
	Memory      int
	TCPPorts    string
	UDPPorts    string
}
