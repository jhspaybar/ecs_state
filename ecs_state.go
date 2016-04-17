// The ecs_state package provides a number of methods to track, update, and query the shared state of an AWS ECS cluster.
// Because ECS exposes the state of the cluster in shared state manner, it is expected for applications monitoring and
// placing tasks within the ECS cluster to replicate the cluster state into a local working copy and synchronize on occassion.
//
// Author: William Thurston
package ecsstate

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

// The State object provides methods to synchronize and query the state of the ECS cluster.
type State struct {
	clusterName string
	db          *gorm.DB
	ecs_client  *ecs.ECS
	log         Logger
}

// Create a new State object.  The clusterName is the cluster to track, ecs_client should be provided by the caller
// with proper credentials preferably scoped to read only access to ECS APIs, and the logger can use ecs_state.DefaultLogger
// for output on stdout, or the user can provide a custom logger instead.
func Initialize(clusterName string, ecs_client *ecs.ECS, logger Logger) StateOps {
	logger.Info("Intializing ecs_state for cluster ", clusterName)

	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		logger.Error("Unable to initialize local database for ecs_state")
		os.Exit(1)
	}

	db.SetLogger(logger)
	db.AutoMigrate(&Cluster{}, &ContainerInstance{}, &Task{}, &TaskDefinition{})
	db.Model(&ContainerInstance{}).AddIndex("idx_remaining_cpu_memory_tcp_udp", "remaining_cpu", "remaining_memory", "remaining_tcp_ports", "remaining_udp_ports")

	return &State{clusterName: clusterName, db: db, ecs_client: ecs_client, log: logger}
}

// Provides direct access to the database through gorm to allow more advanced queries against state.
func (state *State) DB() *gorm.DB {
	return state.db
}

// Will parse and log any AWS errors received while contacting ECS.
func (state *State) handleAwsError(err error) {
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			// Generic AWS error with Code, Message, and original error (if any)
			state.log.Error("Error encountered calling ECS", awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				// A service error occurred
				state.log.Error(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
			}
		} else {
			// This case should never be hit, the SDK should always return an
			// error which satisfies the awserr.Error interface.
			state.log.Error(err.Error())
		}
	}
}

// Many ECS Apis return a generic Failure object, this methods parses and logs generic Failures.
func (state *State) handleFailures(failures []*ecs.Failure) {
	if len(failures) != 0 {
		state.log.Warn("Encountered", len(failures), "failures when contacting ECS")
		for _, failure := range failures {
			state.log.Warn("Failure ARN:", *failure.Arn, ", Reason:", *failure.Reason)
		}
	}
}

// Performs ECS DescribeCluster call on the clusterName provided at Initialization time and updates the local copy of state.
func (state *State) RefreshClusterState() {
	state.log.Info("entering RefreshClusterState()")
	params := &ecs.DescribeClustersInput{
		Clusters: []*string{
			aws.String(state.clusterName),
		},
	}
	resp, err := state.ecs_client.DescribeClusters(params)
	if err != nil {
		state.handleAwsError(err)
		return
	}

	state.handleFailures(resp.Failures)

	for _, cluster := range resp.Clusters {
		clusterModel := Cluster{}
		state.db.Where(Cluster{ARN: *cluster.ClusterArn}).Assign(Cluster{Name: *cluster.ClusterName, Status: *cluster.Status}).FirstOrCreate(&clusterModel)
		state.log.Debug(fmt.Sprintf("Refreshed cluster: %+v", cluster))
	}
}

// Lists and Describes ContainerInstances in the ECS API and stores them in a more queryable form locally.
// Any ContainerInstances no longer returned by ECS, for example if they have been deregistered, will be
// removed from the local view of state as well.
func (state *State) RefreshContainerInstanceState() {
	state.log.Info("entering RefreshContainerInstanceState()")
	params := &ecs.ListContainerInstancesInput{
		Cluster: aws.String(state.clusterName),
	}

	cluster := state.FindClusterByName(state.clusterName)
	refreshTime := int(time.Now().Unix())
	err := state.ecs_client.ListContainerInstancesPages(params, func(page *ecs.ListContainerInstancesOutput, lastPage bool) bool {
		params := &ecs.DescribeContainerInstancesInput{
			ContainerInstances: page.ContainerInstanceArns,
			Cluster:            aws.String(state.clusterName),
		}
		resp, err := state.ecs_client.DescribeContainerInstances(params)
		if err != nil {
			state.handleAwsError(err)
			return !lastPage
		}

		state.handleFailures(resp.Failures)

		for _, containerInstance := range resp.ContainerInstances {
			containerInstanceModel := ContainerInstance{}
			finder := ContainerInstance{
				ARN: *containerInstance.ContainerInstanceArn,
			}
			assignment := state.containerInstanceAssignment(cluster, containerInstance)
			assignment.RefreshTime = refreshTime
			state.db.Where(finder).Assign(assignment).FirstOrCreate(&containerInstanceModel)
			state.log.Debug(fmt.Sprintf("Refreshed ContainerInstance: %+v", containerInstance))
		}

		return !lastPage
	})

	if err != nil {
		state.handleAwsError(err)
		return
	}

	oldContainerInstances := []ContainerInstance{}
	state.DB().Where("refresh_time < ?", refreshTime).Find(&oldContainerInstances)
	state.log.Debug(fmt.Sprintf("Found %d old Container Instances", len(oldContainerInstances)))
	for _, oldContainerInstance := range oldContainerInstances {
		state.DB().Delete(&oldContainerInstance)
	}

}

// Lists and Describes Tasks in the ECS API and stores them in a more queryable form locally.
// Any Tasks no longer returned by ECS, for example if they have been stopped, will be
// removed from the local view of state as well.
func (state *State) RefreshTaskState() {
	params := &ecs.ListTasksInput{
		Cluster: aws.String(state.clusterName),
	}

	refreshTime := int(time.Now().Unix())
	err := state.ecs_client.ListTasksPages(params, func(page *ecs.ListTasksOutput, lastPage bool) bool {
		params := &ecs.DescribeTasksInput{
			Tasks:   page.TaskArns,
			Cluster: aws.String(state.clusterName),
		}
		resp, err := state.ecs_client.DescribeTasks(params)
		if err != nil {
			state.handleAwsError(err)
			return !lastPage
		}

		state.handleFailures(resp.Failures)

		for _, task := range resp.Tasks {
			taskModel := Task{}
			finder := Task{
				ARN: *task.TaskArn,
			}
			assignment := state.taskAssignment(task)
			assignment.RefreshTime = refreshTime
			state.DB().Where(finder).Assign(assignment).FirstOrCreate(&taskModel)
			state.log.Debug(fmt.Sprintf("Refreshed Task: %+v", task))
		}

		return !lastPage
	})

	if err != nil {
		state.handleAwsError(err)
		return
	}

	oldTasks := []Task{}
	state.DB().Where("refresh_time < ?", refreshTime).Find(&oldTasks)
	state.log.Debug(fmt.Sprintf("Found %d old Tasks", len(oldTasks)))
	for _, oldTask := range oldTasks {
		state.DB().Delete(&oldTask)
	}
}

// Creates a Task model to be used in a gorm Assign() call
func (state *State) taskAssignment(task *ecs.Task) Task {
	assignment := Task{
		ClusterARN:           *task.ClusterArn,
		ContainerInstanceARN: *task.ContainerInstanceArn,
		TaskDefinitionARN:    *task.TaskDefinitionArn,
		DesiredStatus:        *task.DesiredStatus,
		LastStatus:           *task.DesiredStatus,
	}
	if task.StartedBy != nil {
		assignment.StartedBy = *task.StartedBy
	}
	return assignment
}

// Unpack a list of ECS resources to retrieve a single resources value as a string, for example the CPU remaining a Container Instance.
func (state *State) getResourceAsInt(resources []*ecs.Resource, name string, defaultValue int) int {
	for _, resource := range resources {
		if *resource.Name == name && *resource.Type == "INTEGER" {
			return int(*resource.IntegerValue)
		}
	}

	return defaultValue
}

// Unpack a list of ECS resources to retrieve the ports still available on a Container Instance
func (state *State) getResourceAsPortSet(resources []*ecs.Resource, name string, defaultValue string) string {
	for _, resource := range resources {
		if *resource.Name == name && *resource.Type == "STRINGSET" {
			return state.portStringBuilder(resource.StringSetValue)
		}
	}

	return defaultValue
}

// A searchable string representation of ports in use to allow for queries of local state with port constraints.
func (state *State) portStringBuilder(ports []*string) string {
	var buffer bytes.Buffer
	for _, port := range ports {
		buffer.WriteString(fmt.Sprintf("=%s=", *port))
	}

	return buffer.String()
}

// Creates a ContainerInstance model to be used in a gorm Assign() call
func (state *State) containerInstanceAssignment(cluster Cluster, containerInstance *ecs.ContainerInstance) ContainerInstance {
	assignment := ContainerInstance{ClusterARN: cluster.ARN}
	if containerInstance.AgentConnected != nil {
		assignment.AgentConnected = *containerInstance.AgentConnected
	}
	if containerInstance.VersionInfo != nil {
		vi := containerInstance.VersionInfo
		if vi.AgentHash != nil {
			assignment.AgentHash = *vi.AgentHash
		}
		if vi.AgentVersion != nil {
			assignment.AgentVersion = *vi.AgentVersion
		}
		if vi.DockerVersion != nil {
			assignment.DockerVersion = *vi.DockerVersion
		}
	}
	if containerInstance.AgentUpdateStatus != nil {
		assignment.AgentUpdateStatus = *containerInstance.AgentUpdateStatus
	}
	if containerInstance.Ec2InstanceId != nil {
		assignment.EC2InstanceId = *containerInstance.Ec2InstanceId
	}
	if containerInstance.RegisteredResources != nil {
		assignment.RegisteredCPU = state.getResourceAsInt(containerInstance.RegisteredResources, "CPU", 0)
		assignment.RegisteredMemory = state.getResourceAsInt(containerInstance.RegisteredResources, "MEMORY", 0)
		assignment.RegisteredTCPPorts = state.getResourceAsPortSet(containerInstance.RegisteredResources, "PORTS", "")
		assignment.RegisteredUDPPorts = state.getResourceAsPortSet(containerInstance.RegisteredResources, "PORTS_UDP", "")
	}
	if containerInstance.RemainingResources != nil {
		assignment.RemainingCPU = state.getResourceAsInt(containerInstance.RemainingResources, "CPU", 0)
		assignment.RemainingMemory = state.getResourceAsInt(containerInstance.RemainingResources, "MEMORY", 0)
		assignment.RemainingTCPPorts = state.getResourceAsPortSet(containerInstance.RemainingResources, "PORTS", "")
		assignment.RemainingUDPPorts = state.getResourceAsPortSet(containerInstance.RemainingResources, "PORTS_UDP", "")
	}
	if containerInstance.Status != nil {
		assignment.Status = *containerInstance.Status
	}
	return assignment
}

// Load the cluster and all ContainerInstances and Tasks into memory as Go objects.
func (state *State) FindClusterByName(name string) Cluster {
	state.log.Info("entering FindClusterByName()")
	cluster := Cluster{}
	state.DB().Where("name = ?", name).Preload("ContainerInstances").Preload("Tasks").Preload("ContainerInstances.Tasks").First(&cluster)
	return cluster
}

// Resolve and cache locally a Task Definition from either a short string like my_app:1 or a full ARN.
func (state *State) FindTaskDefinition(td string) TaskDefinition {
	state.log.Info("entering FindTaskDefinition()")
	queryString := "short_string = ?"
	if strings.HasPrefix(td, "arn:aws:ecs:") {
		queryString = "a_r_n = ?"
	}

	state.log.Debug("Query prefix is:", queryString)
	taskDefinition := TaskDefinition{}
	if state.DB().Where(queryString, td).First(&taskDefinition).RecordNotFound() {
		state.log.Debug(fmt.Sprintf("TaskDefinition %s not found, calling ECS service.", td))
		params := &ecs.DescribeTaskDefinitionInput{
			TaskDefinition: aws.String(td),
		}
		resp, err := state.ecs_client.DescribeTaskDefinition(params)
		if err != nil {
			state.handleAwsError(err)
		}

		taskDefinition = TaskDefinition{
			ARN:         *resp.TaskDefinition.TaskDefinitionArn,
			ShortString: fmt.Sprintf("%s:%s", *resp.TaskDefinition.Family, strconv.Itoa(int(*resp.TaskDefinition.Revision))),
			Cpu:         0,
			Memory:      0,
		}

		tcpPorts := []string{}
		udpPorts := []string{}
		for _, containerDefinition := range resp.TaskDefinition.ContainerDefinitions {
			taskDefinition.Cpu += int(*containerDefinition.Cpu)
			taskDefinition.Memory += int(*containerDefinition.Memory)
			for _, portMapping := range containerDefinition.PortMappings {
				if portMapping.HostPort != nil && *portMapping.HostPort != 0 {
					if portMapping.Protocol != nil && *portMapping.Protocol == ecs.TransportProtocolUdp {
						udpPorts = append(udpPorts, strconv.Itoa(int(*portMapping.HostPort)))
					} else {
						tcpPorts = append(tcpPorts, strconv.Itoa(int(*portMapping.HostPort)))
					}
				}
			}
		}
		taskDefinition.TCPPorts = strings.Join(tcpPorts, ",")
		taskDefinition.UDPPorts = strings.Join(udpPorts, ",")

		state.DB().Create(&taskDefinition)
		state.log.Debug(fmt.Sprintf("Inserted TaskDefinition: %+v", taskDefinition))
	}

	state.log.Debug(fmt.Sprintf("TaskDefinition is: %+v", taskDefinition))
	return taskDefinition
}

// Create a query for port constraints
func (state *State) buildPortQuery(column, ports string) string {
	query := []string{}
	for _, port := range strings.Split(ports, ",") {
		if len(port) == 0 {
			continue
		}
		// instr(a, b) will return zero if column a does not container string b.
		// This format of query matches our serialization and allows for efficient port constraint.
		query = append(query, fmt.Sprintf("instr(%s,\"=%s=\") = 0", column, port))
	}
	return strings.Join(query, " AND ")
}

// Returns all ContainerInstances where the desired TaskDefinition has resources available.
// Additional filtering or constraints can be added if required.
func (state *State) FindLocationsForTaskDefinition(td string) *[]ContainerInstance {
	state.log.Info("entering FindLocationsForTaskDefinition()")
	taskDefinition := state.FindTaskDefinition(td)

	query := []string{"remaining_cpu >= ? AND remaining_memory >= ? AND agent_connected = ?"}
	tcp_query := state.buildPortQuery("remaining_tcp_ports", taskDefinition.TCPPorts)
	if len(tcp_query) > 0 {
		query = append(query, tcp_query)
	}
	udp_query := state.buildPortQuery("remaining_udp_ports", taskDefinition.UDPPorts)
	if len(udp_query) > 0 {
		query = append(query, udp_query)
	}
	fullQuery := strings.Join(query, " AND ")
	state.log.Debug("Full query is:", fullQuery)

	containerInstances := []ContainerInstance{}
	state.DB().Where(fullQuery, taskDefinition.Cpu, taskDefinition.Memory, true).Find(&containerInstances)
	return &containerInstances
}
