package ecsstate

// Local representation of an ECS ContainerInstance and stored by gorm.
// Notably, resources and other sub-objects have been placed into their own
// columns for more robust query capabilities.
type ContainerInstance struct {
	ARN                string `sql:"size:1024" gorm:"primary_key"`
	AgentConnected     bool
	AgentHash          string
	AgentVersion       string
	AgentUpdateStatus  string
	ClusterARN         string `sql:"size:1024;index"`
	DockerVersion      string
	EC2InstanceId      string
	RegisteredCPU      int    `gorm:"column:registered_cpu"`
	RegisteredMemory   int    `gorm:"column:registered_memory"`
	RegisteredTCPPorts string `sql:"size:1024" gorm:"column:registered_tcp_ports"`
	RegisteredUDPPorts string `sql:"size:1024" gorm:"column:registered_udp_ports"`
	RemainingCPU       int    `gorm:"column:remaining_cpu"`
	RemainingMemory    int    `gorm:"column:remaining_memory"`
	RemainingTCPPorts  string `sql:"size:1024" gorm:"column:remaining_tcp_ports"`
	RemainingUDPPorts  string `sql:"size:1024" gorm:"column:remaining_udp_ports"`
	Status             string
	Tasks              []Task

	// Not part of the ECS API
	RefreshTime int
}
