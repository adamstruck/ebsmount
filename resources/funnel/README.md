## Use of ebsmount on AWS Batch with Funnel

[Funnel](https://github.com/ohsu-comp-bio/funnel) is a toolkit for distributed, batch task execution, including a server, worker, and a set of compute, storage, and database backends. 
Given a task description, Funnel will find a worker to execute the task, download inputs, run a series of (Docker) containers, upload outputs, capture logs, and track the whole process.


[AWS Batch](https://aws.amazon.com/documentation/batch/) enables you to run batch computing workloads on the AWS Cloud. Batch computing is a common way for developers, scientists, 
and engineers to access large amounts of compute resources. AWS Batch removes the undifferentiated heavy lifting of configuring and managing the required infrastructure.


AWS Batch tasks, by default, launch the ECS Optimized AMI which includes an 8GB volume for the operating system and a 22GB volume for Docker image and metadata storage. 
The default Docker configuration allocates up to 10GB of this storage to each container instance. [Read more about the default AMI](http://docs.aws.amazon.com/AmazonECS/latest/developerguide/ecs-optimized_AMI.html). 
Due to these limitations, we recommend [creating a custom AMI](http://docs.aws.amazon.com/batch/latest/userguide/create-batch-ami.html). Because AWS Batch has the same requirements for your AMI as 
Amazon ECS, use the default Amazon ECS-optimized Amazon Linux AMI as a base and change it to better suite your tasks.


_ebsmount_ can be packaged into a custom AMI to allow your AWS Batch Jobs to mount an EBS volume per job. Check out this [init.d script](https://github.com/adamstruck/ebsmount/blob/master/resources/init.d/ebsmount) to see
how you can configure _ebsmount_ to run as a service on your custom AMI. 


Check out the example Job Definition and bash script in this folder to see how these services are used together. 
