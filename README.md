Helps manage the state of an ECS cluster to allow for simpler scheduler creation.

This library provides a handy query and state synchronization API so that intelligent
placement of ECS Tasks can be quickly made and tracked.

Example usage:
```
client := ecs.New(&aws.Config{Region: aws.String("us-east-1")})
state := ecs_state.Initialize("default", client, ecs_state.DefaultLogger)
state.RefreshClusterState()
state.RefreshContainerInstanceState()
state.RefreshTaskState()
fmt.Printf("Found Cluster: %+v\n", state.FindClusterByName("default"))
fmt.Printf("Found Locations: %+v\n", state.FindLocationsForTaskDefinition("console-sample-app-static:1"))
```
When run against the "default" cluster created with a single ContainerInstance by the AWS ECS Getting Started Wizard,
you should expect to see the Cluster, ContainerInstance, and Task output, along with an empty array of possible locations
to place the first TaskDefinition created.  No locations are found because of a port conflict.  If you were to scale down
the service created in the Getting Started Wizard to 0, running this code again would yield the now available ContainerInstance
as a location found.

For more details please see http://williamthurston.com/2015/08/20/create-custom-aws-ecs-schedulers-with-ecs-state.html
