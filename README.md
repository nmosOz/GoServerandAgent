# GoServerandAgent
A simple Go server and agent 

## The Server 
### Endpoints 
<li> './all' - This will retrieve everything in DynamoDB and dump it to the screen</li>
<li> './search' - At the current iteration, this endpoint is passed a query (specifically, a bird's common name) and returns a 200 and the information about the sightings of that bird</li>
<li> './status' - This endpoint will display the amount of items in the database and if the database is currently up</li>

This uses a microservice infrastructure 

### <b><ins>What I did!</ins></b>

<li> Built a barebones HTTP server using GoLang </li>
<li> Made server better using GorillaMux </li>
<li> Made rudimentary middleware  </li>
<li> Blocked any non GET requests </li>
<li> Containerized the server application using a multi-stage build Dockerfile for space efficiency due to it being on EC2 </li>
<li> Used DynamoDB capabilities to query the database with GET requests </li>
<li> Added a data validation step that used regular expressions to check if the queries are there and make sense </li>
<li> Used AWS's CodeBuild/CodePipeline for automatic deployment and updating of the microservice </li>
<li> Migrated the Docker container containing the server to AWS ECS and the Application Load Balancer </li>

<br>

<b><ins>Building Container</ins>:<b>
docker build . -t <IMAGE_NAME>

<b><ins>Running Container</ins>:<b>
docker run -e <LOGGLY_TOKEN> -p <DOCKER_PORT:OPENED_PORT> <IMAGE_NAME>
